package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/xbmlz/webber/config"
	"github.com/xbmlz/webber/datasource"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
	config *Config
	logger datasource.Logger
}

type Config struct {
	Driver      string
	Host        string
	Port        int
	Username    string
	Password    string
	Database    string
	Params      string
	MaxOpenConn int
	MaxIdleConn int
}

var errUnsupportedDialect = fmt.Errorf("unsupported db dialect; supported dialects are - mysql, postgres, sqlite")

func New(config config.Config, logger datasource.Logger) *DB {

	dbConfig := getConfig(config)

	// if Hostname is not provided, we won't try to connect to DB
	if dbConfig.Driver != "sqlite" && dbConfig.Host == "" {
		logger.Debugf("skipping database connection initialization as 'DB_HOST' is not provided")
		return nil
	}

	dialector, err := getDialector(dbConfig)
	if err != nil {
		logger.Error(errUnsupportedDialect)
		return nil
	}

	database := &DB{config: dbConfig, logger: logger}

	database.DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		logger.Errorf("failed to connect to database: %v", err)
		return database
	}

	sqlDB, err := database.DB.DB()
	if err != nil {
		logger.Errorf("failed to get sql.DB: %v", err)
	}

	// We are not setting idle connection timeout because we are checking for connection
	// every 10 seconds which would need a connection, moreover if connection expires it is
	// automatically closed by the database/sql package.
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConn)
	// We are not setting max open connection because any connection which is expired,
	// it is closed automatically.
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConn)

	if err := sqlDB.Ping(); err != nil {
		logger.Errorf("failed to ping database: %v", err)
	}

	logger.Debugf("connected to database")

	return database
}

func getDialector(dbConfig *Config) (dialector gorm.Dialector, err error) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database, dbConfig.Params)

	if dbConfig.Driver == "sqlite" {
		// if path is not exist, create it
		if _, err := os.Stat(dbConfig.Database); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(dbConfig.Database), os.ModePerm)
		}
		dsn = fmt.Sprintf("file:%s", dbConfig.Database)
	}

	switch dbConfig.Driver {
	case "sqlite":
		dialector = sqlite.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		return nil, errUnsupportedDialect
	}
	return dialector, nil
}

func getConfig(c config.Config) *Config {
	maxIdleConn, _ := c.GetInt("DB_MAX_IDLE_CONNECTION", 2)

	maxOpenConn, _ := c.GetInt("DB_MAX_OPEN_CONNECTION", 0)

	port, _ := c.GetInt("DB_PORT", 0)

	return &Config{
		Driver:      c.GetString("DB_DRIVER", ""),
		Host:        c.GetString("DB_HOST", ""),
		Password:    c.GetString("DB_PASSWORD", ""),
		Database:    c.GetString("DB_NAME", ""),
		Params:      c.GetString("DB_PARAMS", ""),
		MaxOpenConn: maxOpenConn,
		MaxIdleConn: maxIdleConn,
		Port:        port,
	}
}
