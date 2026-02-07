package zap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/robert-pkg/base4go/log"
)

type Options struct {
	log.Options
}

type configKey struct{}

// WithConfig pass zap.Config to logger.
func WithConfig(c zap.Config) log.Option {
	return log.SetOption(configKey{}, c)
}

type encoderConfigKey struct{}

// WithEncoderConfig pass zapcore.EncoderConfig to logger.
func WithEncoderConfig(c zapcore.EncoderConfig) log.Option {
	return log.SetOption(encoderConfigKey{}, c)
}

type namespaceKey struct{}

func WithNamespace(namespace string) log.Option {
	return log.SetOption(namespaceKey{}, namespace)
}

type optionsKey struct{}

func WithOptions(opts ...zap.Option) log.Option {
	return log.SetOption(optionsKey{}, opts)
}

type loggerKey struct{}

// WithLogger pass zap.Logger to logger
func WithLogger(zapLogger *zap.Logger) log.Option {
	return log.SetOption(loggerKey{}, zapLogger)
}
