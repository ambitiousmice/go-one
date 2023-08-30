package log

import (
	"gopkg.in/yaml.v2"
	"net/url"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapConfig struct {
	Level             string                 `default:"info" json:"level,omitempty" yaml:"level" property:"level"`
	Development       bool                   `default:"false" json:"development,omitempty" yaml:"development" property:"development"`
	DisableCaller     bool                   `default:"false" json:"disable-caller,omitempty" yaml:"disable-caller" property:"disable-caller"`
	DisableStacktrace bool                   `default:"false" json:"disable-stacktrace,omitempty" yaml:"disable-stacktrace" property:"disable-stacktrace"`
	Encoding          string                 `default:"console" json:"encoding,omitempty" yaml:"encoding" property:"encoding"`
	EncoderConfig     EncoderConfig          `default:"" json:"encoder-config,omitempty" yaml:"encoder-config" property:"encoder-config"`
	OutputPaths       []string               `default:"[\"stderr\"]" json:"output-paths,omitempty" yaml:"output-paths" property:"output-paths"`
	ErrorOutputPaths  []string               `default:"[\"stderr\"]" json:"error-output-paths,omitempty" yaml:"error-output-paths" property:"error-output-paths"`
	InitialFields     map[string]interface{} `default:"" json:"initial-fields,omitempty" yaml:"initial-fields" property:"initial-fields"`
}

type LoggerConfig struct {
	LumberjackConfig *lumberjack.Logger `yaml:"lumberjack-config" json:"lumberjack-config,omitempty" property:"lumberjack-config"`
	ZapConfig        ZapConfig          `yaml:"zap-config" json:"zap-config,omitempty" property:"zap-config"`
}

type EncoderConfig struct {
	MessageKey     string            `default:"message" json:"message-key,omitempty" yaml:"message-key" property:"message-key"`
	LevelKey       string            `default:"level" json:"level-key,omitempty" yaml:"level-key" property:"level-key"`
	TimeKey        string            `default:"time" json:"time-key,omitempty" yaml:"time-key" property:"time-key"`
	NameKey        string            `default:"logger" json:"name-key,omitempty" yaml:"name-key" property:"name-key"`
	CallerKey      string            `default:"caller" json:"caller-key,omitempty" yaml:"caller-key" property:"caller-key"`
	StacktraceKey  string            `default:"stacktrace" json:"stacktrace-key,omitempty" yaml:"stacktrace-key" property:"stacktrace-key"`
	EncodeLevel    string            `default:"capitalColor" json:"level-encoder" yaml:"level-encoder" property:"level-encoder"`
	EncodeTime     string            `default:"iso8601" json:"time-encoder" yaml:"time-encoder" property:"time-encoder"`
	EncodeDuration string            `default:"seconds" json:"duration-encoder" yaml:"duration-encoder" property:"duration-encoder"`
	EncodeCaller   string            `default:"short" json:"caller-encoder" yaml:"calle-encoder" property:"caller-encoder"`
	Params         map[string]string `yaml:"params" json:"params,omitempty"`
}

func (LoggerConfig) Prefix() string {
	return "logger"
}

func (lc *LoggerConfig) Init() error {
	bytes, err := yaml.Marshal(lc)
	if err != nil {
		return err
	}

	logConf := &Config{}
	if err = yaml.Unmarshal(bytes, logConf); err != nil {
		return err
	}
	err = lc.ZapConfig.EncoderConfig.setEncoderConfig(&(logConf.ZapConfig.EncoderConfig))
	if err != nil {
		return err
	}
	lc.ZapConfig.setZapConfig(logConf.ZapConfig)
	InitLogger(logConf)
	return nil
}

func (e *ZapConfig) setZapConfig(config *zap.Config) {
	config.OutputPaths = e.OutputPaths
	config.ErrorOutputPaths = e.ErrorOutputPaths
	config.DisableStacktrace = e.DisableStacktrace
	config.DisableCaller = e.DisableCaller
	config.InitialFields = e.InitialFields
}

func (e *EncoderConfig) setEncoderConfig(encoderConfig *zapcore.EncoderConfig) error {
	encoderConfig.MessageKey = e.MessageKey
	encoderConfig.LevelKey = e.LevelKey
	encoderConfig.TimeKey = e.TimeKey
	encoderConfig.NameKey = e.NameKey
	encoderConfig.CallerKey = e.CallerKey
	encoderConfig.StacktraceKey = e.StacktraceKey

	if err := encoderConfig.EncodeLevel.UnmarshalText([]byte(e.EncodeLevel)); err != nil {
		return err
	}

	if err := encoderConfig.EncodeTime.UnmarshalText([]byte(e.EncodeTime)); err != nil {
		return err
	}

	if err := encoderConfig.EncodeDuration.UnmarshalText([]byte(e.EncodeDuration)); err != nil {
		return err
	}

	if err := encoderConfig.EncodeCaller.UnmarshalText([]byte(e.EncodeCaller)); err != nil {
		return err
	}
	return nil
}

func (lc *LoggerConfig) getUrlMap() url.Values {
	urlMap := url.Values{}
	for key, val := range lc.ZapConfig.EncoderConfig.Params {
		urlMap.Set(key, val)
	}
	return urlMap
}

type LoggerConfigBuilder struct {
	loggerConfig *LoggerConfig
}

func NewLoggerConfigBuilder() *LoggerConfigBuilder {
	return &LoggerConfigBuilder{loggerConfig: &LoggerConfig{}}
}

func (lcb *LoggerConfigBuilder) SetLumberjackConfig(lumberjackConfig *lumberjack.Logger) *LoggerConfigBuilder {
	lcb.loggerConfig.LumberjackConfig = lumberjackConfig
	return lcb
}

func (lcb *LoggerConfigBuilder) SetZapConfig(zapConfig ZapConfig) *LoggerConfigBuilder {
	lcb.loggerConfig.ZapConfig = zapConfig
	return lcb
}

func (lcb *LoggerConfigBuilder) Build() *LoggerConfig {
	return lcb.loggerConfig
}

// DynamicUpdateProperties dynamically update properties.
func (lc *LoggerConfig) DynamicUpdateProperties(newLoggerConfig *LoggerConfig) {
	if newLoggerConfig != nil && lc.ZapConfig.Level != newLoggerConfig.ZapConfig.Level {
		lc.ZapConfig.Level = newLoggerConfig.ZapConfig.Level
		logger.Infof("LoggerConfig's ZapConfig Level was dynamically updated, new value:%v", lc.ZapConfig.Level)
	}
}
