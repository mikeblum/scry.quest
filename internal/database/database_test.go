package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mikeblum/scry.quest/conf"
)

func pgTestConf(ctx context.Context) (Config, error) {
	config, err := conf.New(ctx, "")
	if err != nil {
		return Config{}, err
	}

	return Config{
		Host:     getConfigValue(config, "scry.postgres.host", "localhost"),
		Port:     getConfigValue(config, "scry.postgres.port", "5432"),
		User:     getConfigValue(config, "scry.postgres.user", "postgres"),
		Password: getConfigValue(config, "scry.postgres.password", ""),
		Database: getConfigValue(config, "scry.postgres.database", "postgres"),
		SSLMode:  getConfigValue(config, "scry.postgres.sslmode", "disable"),
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
	db     *Database
	ctx    context.Context
	config Config
}

func (suite *DatabaseTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	config, err := pgTestConf(suite.ctx)
	if err != nil {
		suite.T().Skip("Failed to load test configuration:", err)
	}
	suite.config = config

	db, err := NewDatabase(suite.ctx, config)
	if err != nil {
		suite.T().Skip("PostgreSQL not available for testing:", err)
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
