package log

import (
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/xbmlz/webber/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelFatal = "fatal"
	LevelPanic = "panic"
)

const (
	defaultLevel      = LevelInfo
	defaultMaxBackups = 3
	defaultMaxSize    = 10
	defaultMaxAge     = 7
	defaultCompress   = true
	defaultEncoder    = "console"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	GetLogger() *zap.Logger
}

type Config struct {
	Level      string // env var: LOG_LEVEL
	File       string // env var: LOG_FILE
	Encoder    string // env var: LOG_ENCODER (json, console)
	MaxBackups int    // env var: LOG_MAX_BACKUPS
	MaxSize    int    // env var: LOG_MAX_SIZE
	MaxAge     int    // env var: LOG_MAX_AGE
	Compress   bool   // env var: LOG_COMPRESS
}

type logger struct {
	config *Config
	logger *zap.Logger
}

func New(level string) Logger {
	logger := &logger{}
	logger.initZapLogger(level, "console")
	return logger
}

func NewWithConfg(cfg config.Config) Logger {
	logger := &logger{}
	logger.loadConfig(cfg)
	logger.initZapLogger(logger.config.Level, logger.config.Encoder)
	return logger
}

func (l *logger) GetLogger() *zap.Logger {
	return l.logger
}

func (l *logger) initZapLogger(level, encoder string) {
	cores := []zapcore.Core{
		l.getConsoleCore(level, encoder),
	}
	if l.config != nil && l.config.File != "" {
		// create file path if not exists
		if _, err := os.Stat(l.config.File); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(l.config.File), os.ModePerm)
		}
		cores = append(cores, l.getFileCore())
	}
	zapOpts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}
	if level == "debug" {
		zapOpts = append(zapOpts, zap.Development(), zap.AddStacktrace(zapcore.ErrorLevel))
	}

	logger := zap.New(zapcore.NewTee(cores...), zapOpts...)

	defer logger.Sync()

	l.logger = logger
}

func (l *logger) getConsoleCore(level, encoder string) zapcore.Core {
	consoleEncoderConfig := zap.NewProductionEncoderConfig()
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	consoleEncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	consoleEncoderConfig.EncodeName = zapcore.FullNameEncoder
	consoleEncoderConfig.ConsoleSeparator = "\t"
	var consoleEncoder zapcore.Encoder

	if encoder == "console" {
		consoleEncoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
	} else {
		consoleEncoder = zapcore.NewJSONEncoder(consoleEncoderConfig)
	}

	return zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(colorable.NewColorableStdout()),
		zap.NewAtomicLevelAt(ParseLevel(level)),
	)
}

func (l *logger) getFileCore() zapcore.Core {
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	fileEncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	fileEncoderConfig.EncodeName = zapcore.FullNameEncoder
	fileEncoderConfig.ConsoleSeparator = "\t"
	var fileEncoder zapcore.Encoder

	if l.config.Encoder == "json" {
		fileEncoder = zapcore.NewJSONEncoder(fileEncoderConfig)
	} else {
		fileEncoder = zapcore.NewConsoleEncoder(fileEncoderConfig)
	}

	hook := &lumberjack.Logger{
		Filename:   l.config.File,
		MaxSize:    l.config.MaxSize,
		MaxBackups: l.config.MaxBackups,
		MaxAge:     l.config.MaxAge,
		Compress:   l.config.Compress,
	}

	return zapcore.NewCore(
		fileEncoder,
		zapcore.AddSync(hook),
		zap.NewAtomicLevelAt(ParseLevel(l.config.Level)),
	)
}

func (l *logger) loadConfig(cfg config.Config) {
	maxAge, _ := cfg.GetInt("LOG_MAX_AGE", defaultMaxAge)
	maxBackups, _ := cfg.GetInt("LOG_MAX_BACKUPS", defaultMaxBackups)
	maxSize, _ := cfg.GetInt("LOG_MAX_SIZE", defaultMaxSize)
	compress, _ := cfg.GetBool("LOG_COMPRESS", defaultCompress)
	l.config = &Config{
		Level:      cfg.GetString("LOG_LEVEL", defaultLevel),
		File:       cfg.GetString("LOG_FILE", ""),
		Encoder:    cfg.GetString("LOG_ENCODER", defaultEncoder),
		MaxBackups: maxBackups,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		Compress:   compress,
	}
}

func (l *logger) Debug(args ...interface{}) {
	l.logger.Sugar().Debug(args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.logger.Sugar().Debugf(format, args...)
}

func (l *logger) Info(args ...interface{}) {
	l.logger.Sugar().Info(args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.logger.Sugar().Infof(format, args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.logger.Sugar().Warn(args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.logger.Sugar().Warnf(format, args...)
}

func (l *logger) Error(args ...interface{}) {
	l.logger.Sugar().Error(args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.logger.Sugar().Errorf(format, args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.logger.Sugar().Fatal(args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.logger.Sugar().Fatalf(format, args...)
}
