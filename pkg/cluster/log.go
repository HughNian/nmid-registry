package cluster

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"nmid-registry/pkg/option"
	"nmid-registry/pkg/utils"
	"path/filepath"
)

func ClientLoggerConfig(opt *option.Options, fileName string) *zap.Config {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "", // no need
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "", // no need
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	level := zap.NewAtomicLevel()
	if opt.ClusterDebug {
		level.SetLevel(zapcore.DebugLevel)
	} else {
		level.SetLevel(zapcore.InfoLevel)
	}

	outputPaths := []string{utils.GOOSPath(filepath.Join(opt.AbsLogDir, fileName))}

	return &zap.Config{
		Level:            level,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: outputPaths,
	}
}
