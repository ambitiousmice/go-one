package log

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var logger Logger

func init() {
	InitLogger(nil)
}

type OneLogger struct {
	Logger
	dynamicLevel zap.AtomicLevel
}

type Config struct {
	LumberjackConfig *lumberjack.Logger `yaml:"lumberjack-config"`
	ZapConfig        *zap.Config        `yaml:"zap-config"`
	CallerSkip       int
}

// Logger is the interface for Logger types
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	Debugf(fmt string, args ...interface{})
	Infof(fmt string, args ...interface{})
	Warnf(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})
	Fatalf(fmt string, args ...interface{})
	Panicf(fmt string, args ...interface{})
}

// InitLogger use for init logger by @conf
func InitLogger(conf *Config) {
	var (
		zapLogger *zap.Logger
		config    = &Config{}
	)
	if conf == nil || conf.ZapConfig == nil {
		zapLoggerEncoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		config.ZapConfig = &zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:      false,
			Encoding:         "console",
			EncoderConfig:    zapLoggerEncoderConfig,
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
	} else {
		config.ZapConfig = conf.ZapConfig
	}

	if conf != nil {
		config.CallerSkip = conf.CallerSkip
	}

	if config.CallerSkip == 0 {
		config.CallerSkip = 1
	}

	if conf == nil || conf.LumberjackConfig == nil {
		zapLogger, _ = config.ZapConfig.Build(zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	} else {
		config.LumberjackConfig = conf.LumberjackConfig
		zapLogger = initZapLoggerWithSyncer(config)
	}

	logger = &OneLogger{Logger: zapLogger.Sugar(), dynamicLevel: config.ZapConfig.Level}
}

// SetLogger sets logger for dubbo and getty
func SetLogger(log Logger) {
	logger = log
}

// GetLogger gets the logger
func GetLogger() Logger {
	return logger
}

// SetLoggerLevel use for set logger level
func SetLoggerLevel(level string) bool {
	if l, ok := logger.(OpsLogger); ok {
		l.SetLoggerLevel(level)
		return true
	}
	return false
}

// OpsLogger use for the SetLoggerLevel
type OpsLogger interface {
	Logger
	SetLoggerLevel(level string)
}

// SetLoggerLevel use for set logger level
func (o *OneLogger) SetLoggerLevel(level string) {
	l := new(zapcore.Level)
	if err := l.Set(level); err == nil {
		o.dynamicLevel.SetLevel(*l)
	}
}

// initZapLoggerWithSyncer init zap Logger with syncer
func initZapLoggerWithSyncer(conf *Config) *zap.Logger {
	core := zapcore.NewCore(
		conf.getEncoder(),
		conf.getLogWriter(),
		zap.NewAtomicLevelAt(conf.ZapConfig.Level.Level()),
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(conf.CallerSkip))
}

// getEncoder get encoder by config, zapcore support json and console encoder
func (c *Config) getEncoder() zapcore.Encoder {
	if c.ZapConfig.Encoding == "json" {
		return zapcore.NewJSONEncoder(c.ZapConfig.EncoderConfig)
	} else if c.ZapConfig.Encoding == "console" {
		return zapcore.NewConsoleEncoder(c.ZapConfig.EncoderConfig)
	}
	return nil
}

// getLogWriter get Lumberjack writer by LumberjackConfig
func (c *Config) getLogWriter() zapcore.WriteSyncer {
	writeSyncer := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(c.LumberjackConfig),
		zapcore.AddSync(os.Stdout),
	)
	return writeSyncer
}
