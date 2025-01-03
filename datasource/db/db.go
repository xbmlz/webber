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
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
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
	LogLevel    string
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

	database.DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger.Default.LogMode(parseLogLevel(dbConfig.LogLevel)),
	})
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

func parseLogLevel(level string) gormLogger.LogLevel {
	switch level {
	case "silent":
		return gormLogger.Silent
	case "error":
		return gormLogger.Error
	case "warn":
		return gormLogger.Warn
	case "info":
		return gormLogger.Info
	default:
		return gormLogger.Info
	}
}

func getDialector(dbConfig *Config) (dialector gorm.Dialector, err error) {

	switch dbConfig.Driver {
	case "sqlite":
		if _, err := os.Stat(dbConfig.Database); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(dbConfig.Database), os.ModePerm)
		}
		dialector = sqlite.Open(fmt.Sprintf("file:%s", dbConfig.Database))
	case "mysql":
		dialector = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database, dbConfig.Params))
	case "postgres":
		dialector = postgres.Open(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s", dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password, dbConfig.Database, dbConfig.Params))
	case "sqlserver":
		dialector = sqlserver.Open(fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&%s", dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database, dbConfig.Params))
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
		Username:    c.GetString("DB_USERNAME", ""),
		Password:    c.GetString("DB_PASSWORD", ""),
		Database:    c.GetString("DB_NAME", ""),
		Params:      c.GetString("DB_PARAMS", ""),
		LogLevel:    c.GetString("DB_LOG_LEVEL", "info"),
		MaxOpenConn: maxOpenConn,
		MaxIdleConn: maxIdleConn,
		Port:        port,
	}
}
