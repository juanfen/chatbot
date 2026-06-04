package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a zap logger.
func New(level string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level = parseLevel(level)
	cfg.Encoding = "console"
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	l, err := cfg.Build(zap.AddCaller())
	if err != nil {
		return zap.NewNop()
	}
	return l
}

func parseLevel(level string) zap.AtomicLevel {
	al := zap.NewAtomicLevel()
	switch level {
	case "debug":
		al.SetLevel(zap.DebugLevel)
	case "warn":
		al.SetLevel(zap.WarnLevel)
	case "error":
		al.SetLevel(zap.ErrorLevel)
	default:
		al.SetLevel(zap.InfoLevel)
	}
	return al
}

func Err(err error) zap.Field { return zap.Error(err) }

func String(key, val string) zap.Field { return zap.String(key, val) }

