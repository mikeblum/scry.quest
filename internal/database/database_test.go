package database

import (
	"context"
	"embed"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mikeblum/scry.quest/conf"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func pgTestConf(ctx context.Context) (Config, error) {
	// Load configuration using conf package which handles .env files and environment variables
	config, err := conf.New(ctx, "")
	if err != nil {
		return Config{}, err
	}

	// Extract database configuration with fallbacks for test environment
	return Config{
		Host:     getConfigValue(config, "postgres.host", "localhost"),
		Port:     getConfigValue(config, "postgres.port", "5432"),
		User:     getConfigValue(config, "postgres.user", "postgres"),
		Password: getConfigValue(config, "postgres.password", ""),
		Database: getConfigValue(config, "postgres.database", "postgres"),
		SSLMode:  getConfigValue(config, "postgres.sslmode", "disable"),
	}, nil
}

func getConfigValue(config *conf.Config, key, fallback string) string {
	if config.Exists(key) {
		return config.String(key)
	}
	return fallback
}

type DatabaseTestSuite struct {
	suite.Suite
	db  *Database
	ctx context.Context
}

func (suite *DatabaseTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	config, err := pgTestConf(suite.ctx)
	if err != nil {
		suite.T().Skip("Failed to load test configuration:", err)
	}

	db, err := NewDatabase(suite.ctx, config)
	if err != nil {
		suite.T().Skip("PostgreSQL not available for testing:", err)
	}
	suite.db = db

	// Run migrations before all tests
	err = suite.db.RunMigrations(suite.ctx, migrationsFS)
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
	// Test that migrations table was created
	var exists bool
	err := suite.db.conn.QueryRow(suite.ctx, "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = 'scry_quest' AND table_name = 'schema_migrations')").Scan(&exists)
	suite.Require().NoError(err)
	suite.True(exists)

	// Test that our main tables were created
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
