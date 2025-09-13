package database

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	charset                      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
	daemonSocket                 = "/tmp/podman.sock"
	envDockerHost                = "DOCKER_HOST"
	envTestContainerRyukDisabled = "TESTCONTAINERS_RYUK_DISABLED"
	postgresImage                = "pgvector/pgvector:pg17"
	sslMode                      = "disable"
	testDB                       = "scry_dev_test"
	testDBUser                   = "postgres"
	testDBPasswdStr              = 16
)

type DatabaseTestSuite struct {
	suite.Suite
	db        *Database
	ctx       context.Context
	config    Config
	container *postgres.PostgresContainer
}

func (suite *DatabaseTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	setupTestContainers(suite.T())

	// generate testing password
	var testDBPasswd string
	var err error
	if testDBPasswd, err = generatePassword(testDBPasswdStr); err != nil {
		suite.Fail("error generating test postgres passwd", "err", err)
	}

	container, err := postgres.Run(suite.ctx,
		postgresImage,
		postgres.WithDatabase(testDB),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPasswd),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_PASSWORD":    testDBPasswd,
			"POSTGRES_INITDB_ARGS": "--auth-host=scram-sha-256",
		}),
		testcontainers.WithCmd(
			testDBUser,
			"-c", "shared_buffers=256MB",
			"-c", "maintenance_work_mem=64MB",
			"-c", "work_mem=4MB",
		),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		suite.T().Fatalf("Failed to start postgres container: %v", err)
	}
	suite.container = container

	host, err := container.Host(suite.ctx)
	if err != nil {
		suite.T().Fatalf("Failed to get container host: %v", err)
	}

	port, err := container.MappedPort(suite.ctx, "5432")
	if err != nil {
		suite.T().Fatalf("Failed to get container port: %v", err)
	}

	suite.config = Config{
		Host:     host,
		Port:     port.Port(),
		User:     testDBUser,
		Password: testDBPasswd,
		Database: testDB,
		SSLMode:  sslMode,
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		suite.config.Host, suite.config.Port, suite.config.User, suite.config.Password, suite.config.Database, suite.config.SSLMode)

	conn, err := pgx.Connect(suite.ctx, connStr)
	if err != nil {
		suite.T().Fatalf("Failed to connect to test database: %v", err)
	}
	defer func() { _ = conn.Close(suite.ctx) }()

	_, err = conn.Exec(suite.ctx, "CREATE USER scry_quest WITH PASSWORD 'scry_quest_test'")
	if err != nil {
		suite.T().Fatalf("Failed to create scry_quest user: %v", err)
	}

	db, err := NewDatabase(suite.ctx, suite.config)
	if err != nil {
		suite.T().Fatalf("PostgreSQL not available for testing: %v", err)
	}
	suite.db = db
}

func (suite *DatabaseTestSuite) SetupTest() {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		suite.config.Host, suite.config.Port, suite.config.User, suite.config.Password, suite.config.Database, suite.config.SSLMode)

	conn, err := pgx.Connect(suite.ctx, connStr)
	if err != nil {
		suite.T().Fatalf("Failed to connect to test database: %v", err)
	}
	defer func() { _ = conn.Close(suite.ctx) }()

	_, err = conn.Exec(suite.ctx, "DROP TABLE IF EXISTS migrations CASCADE")
	if err != nil {
		suite.T().Fatalf("Failed to clean migrations table: %v", err)
	}

	_, err = conn.Exec(suite.ctx, "DROP SCHEMA IF EXISTS scry_quest CASCADE")
	if err != nil {
		suite.T().Fatalf("Failed to clean schema: %v", err)
	}

	err = suite.db.RunMigrations(suite.ctx, suite.config)
	suite.Require().NoError(err)
}

func (suite *DatabaseTestSuite) TearDownSuite() {
	if suite.db != nil {
		_ = suite.db.Close(suite.ctx)
	}
	if suite.container != nil {
		_ = suite.container.Terminate(suite.ctx)
	}
}

func (suite *DatabaseTestSuite) TestDatabaseConnection() {
	suite.NotNil(suite.db.Queries())
	suite.NotNil(suite.db.Conn())
}

func (suite *DatabaseTestSuite) TestMigrationsRan() {
	var exists bool
	err := suite.db.conn.QueryRow(suite.ctx, "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'migrations')").Scan(&exists)
	suite.Require().NoError(err)
	suite.True(exists)

	tables := []string{"spells", "bestiary", "classes", "species"}
	for _, table := range tables {
		err := suite.db.conn.QueryRow(suite.ctx, "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = 'scry_quest' AND table_name = $1)", table).Scan(&exists)
		suite.Require().NoError(err)
		suite.True(exists, "Table %s should exist", table)
	}
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}

func setupTestContainers(t *testing.T) {
	// Configure testcontainers to use Podman if available
	if os.Getenv(envDockerHost) == "" {
		if _, err := os.Stat(daemonSocket); err == nil {
			assert.NoError(t, os.Setenv(envDockerHost, "unix://"+daemonSocket))
		}
	}

	// Disable Ryuk for Podman compatibility
	if os.Getenv(envTestContainerRyukDisabled) == "" {
		assert.NoError(t, os.Setenv(envTestContainerRyukDisabled, "true"))
	}
}

// GeneratePassword generates a cryptographically secure random password
// of the specified length, using a defined set of characters.
func generatePassword(length int) (string, error) {
	password := make([]byte, length)
	_, err := rand.Read(password)
	if err != nil {
		return "", fmt.Errorf("error reading from crypto/rand: %w", err)
	}

	for i := 0; i < length; i++ {
		// Map the random byte to an index within the charset
		password[i] = charset[int(password[i])%len(charset)]
	}

	return string(password), nil
}
