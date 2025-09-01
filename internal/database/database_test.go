package database

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/suite"

	"github.com/mikeblum/scry.quest/conf"
)

func pgTestConf(ctx context.Context) (Config, error) {
	testConf := "../../.env.test"
	config, err := conf.New(ctx, &testConf)
	if err != nil {
		return Config{}, err
	}

	databaseURL := config.GetPrefixedEnv("DATABASE_URL", "postgres://localhost/scry_quest_test?sslmode=disable")

	dbConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse database URL: %w", err)
	}

	return Config{
		Host:     dbConfig.Host,
		Port:     fmt.Sprintf("%d", dbConfig.Port),
		User:     dbConfig.User,
		Password: dbConfig.Password,
		Database: dbConfig.Database,
		SSLMode:  "disable",
	}, nil
}

type DatabaseTestSuite struct {
	suite.Suite
	db     *Database
	ctx    context.Context
	config Config
}

func (suite *DatabaseTestSuite) ensureTestDatabase() error {
	adminConfig := suite.config
	adminConfig.Database = "postgres"

	conn, err := pgx.Connect(suite.ctx, fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		adminConfig.Host,
		adminConfig.Port,
		adminConfig.User,
		adminConfig.Password,
		adminConfig.Database,
		adminConfig.SSLMode,
	))
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer func() {
		_ = conn.Close(suite.ctx)
	}()

	var exists bool
	err = conn.QueryRow(suite.ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", suite.config.Database).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if exists {
		return nil
	}

	// create test db
	dbName := strings.ReplaceAll(suite.config.Database, "'", "''")
	_, err = conn.Exec(suite.ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}

	userConn, err := pgx.Connect(suite.ctx, fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		suite.config.Host,
		suite.config.Port,
		suite.config.User,
		suite.config.Password,
		suite.config.Database,
		suite.config.SSLMode,
	))
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}
	defer func() {
		_ = userConn.Close(suite.ctx)
	}()

	_, err = userConn.Exec(suite.ctx, fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", suite.config.User, suite.config.Password))
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create test user: %w", err)
	}

	_, err = userConn.Exec(suite.ctx, fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", dbName, suite.config.User))
	if err != nil {
		return fmt.Errorf("failed to grant privileges to test user: %w", err)
	}

	return nil
}

func (suite *DatabaseTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	config, err := pgTestConf(suite.ctx)
	if err != nil {
		suite.T().Fatalf("Failed to load test configuration: %v", err)
	}
	suite.config = config

	if err := suite.ensureTestDatabase(); err != nil {
		suite.T().Fatalf("Failed to ensure test database exists: %v", err)
	}

	db, err := NewDatabase(suite.ctx, config)
	if err != nil {
		suite.T().Fatalf("PostgreSQL not available for testing: %v", err)
	}
	suite.db = db

	err = suite.db.RunMigrations(suite.ctx, suite.config)
	suite.Require().NoError(err)
}

func (suite *DatabaseTestSuite) TearDownSuite() {
	if suite.db != nil {
		_ = suite.db.Close(suite.ctx)
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
