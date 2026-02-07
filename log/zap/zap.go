package zap

import (
	"context"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/robert-pkg/base4go/log"
)

type zaplog struct {
	cfg  zap.Config
	zap  *zap.Logger
	opts log.Options
	sync.RWMutex
	fields map[string]interface{}
}

func (l *zaplog) Init(opts ...log.Option) error {

	for _, o := range opts {
		o(&l.opts)
	}

	zapConfig := zap.NewProductionConfig()
	if zconfig, ok := l.opts.Context.Value(configKey{}).(zap.Config); ok {
		zapConfig = zconfig
	}

	zapConfig.Level.SetLevel(loggerToZapLevel(l.opts.Level))

	if zcconfig, ok := l.opts.Context.Value(encoderConfigKey{}).(zapcore.EncoderConfig); ok {
		zapConfig.EncoderConfig = zcconfig
	}

	l.cfg = zapConfig

	// Using zap.Logger instance to replace internal logger
	if zapLogger, ok := l.opts.Context.Value(loggerKey{}).(*zap.Logger); ok && zapLogger != nil {
		l.zap = zapLogger
	} else {
		log, err := zapConfig.Build(
			zap.AddCallerSkip(l.opts.CallerSkipCount),
		)
		if err != nil {
			return err
		}

		l.zap = log
	}

	// Adding seed fields if exist
	if l.opts.Fields != nil {
		data := []zap.Field{}
		for k, v := range l.opts.Fields {
			data = append(data, zap.Any(k, v))
		}
		l.zap = l.zap.With(data...)
	}

	// Adding namespace
	if namespace, ok := l.opts.Context.Value(namespaceKey{}).(string); ok {
		l.zap = l.zap.With(zap.Namespace(namespace))
	}

	// Adding options
	if options, ok := l.opts.Context.Value(optionsKey{}).([]zap.Option); ok {
		l.zap = l.zap.WithOptions(options...)
	}

	// defer log.Sync() ??

	l.fields = make(map[string]interface{})

	return nil
}

func (l *zaplog) Fields(fields map[string]interface{}) log.Logger {
	l.Lock()
	nfields := make(map[string]interface{}, len(l.fields))
	for k, v := range l.fields {
		nfields[k] = v
	}
	l.Unlock()
	for k, v := range fields {
		nfields[k] = v
	}

	data := make([]zap.Field, 0, len(nfields))
	for k, v := range fields {
		data = append(data, zap.Any(k, v))
	}

	zl := &zaplog{
		cfg:    l.cfg,
		zap:    l.zap.With(data...),
		opts:   l.opts,
		fields: make(map[string]interface{}),
	}

	return zl
}

func (l *zaplog) Error(err error) log.Logger {
	return l.Fields(map[string]interface{}{"error": err})
}

func (l *zaplog) Log(level log.Level, args ...interface{}) {
	l.RLock()
	data := make([]zap.Field, 0, len(l.fields))
	for k, v := range l.fields {
		data = append(data, zap.Any(k, v))
	}
	l.RUnlock()

	lvl := loggerToZapLevel(level)
	msg := fmt.Sprint(args...)
	switch lvl {
	case zap.DebugLevel:
		l.zap.Debug(msg, data...)
	case zap.InfoLevel:
		l.zap.Info(msg, data...)
	case zap.WarnLevel:
		l.zap.Warn(msg, data...)
	case zap.ErrorLevel:
		l.zap.Error(msg, data...)
	case zap.FatalLevel:
		l.zap.Fatal(msg, data...)
	}
}

func (l *zaplog) Logf(level log.Level, format string, args ...interface{}) {
	l.RLock()
	data := make([]zap.Field, 0, len(l.fields))
	for k, v := range l.fields {
		data = append(data, zap.Any(k, v))
	}
	l.RUnlock()

	lvl := loggerToZapLevel(level)
	msg := fmt.Sprintf(format, args...)
	switch lvl {
	case zap.DebugLevel:
		l.zap.Debug(msg, data...)
	case zap.InfoLevel:
		l.zap.Info(msg, data...)
	case zap.WarnLevel:
		l.zap.Warn(msg, data...)
	case zap.ErrorLevel:
		l.zap.Error(msg, data...)
	case zap.FatalLevel:
		l.zap.Fatal(msg, data...)
	}
}

func (l *zaplog) String() string {
	return "zap"
}

func (l *zaplog) Options() log.Options {
	return l.opts
}

// New builds a new logger based on options.
func NewLogger(opts ...log.Option) (log.Logger, error) {
	// Default options
	options := log.Options{
		Level:           log.InfoLevel,
		Fields:          make(map[string]interface{}),
		Out:             os.Stderr,
		Context:         context.Background(),
		CallerSkipCount: 2,
	}

	l := &zaplog{opts: options}
	if err := l.Init(opts...); err != nil {
		return nil, err
	}

	return l, nil
}

func loggerToZapLevel(level log.Level) zapcore.Level {
	switch level {
	case log.TraceLevel, log.DebugLevel:
		return zap.DebugLevel
	case log.InfoLevel:
		return zap.InfoLevel
	case log.WarnLevel:
		return zap.WarnLevel
	case log.ErrorLevel:
		return zap.ErrorLevel
	case log.FatalLevel:
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func zapToLoggerLevel(level zapcore.Level) log.Level {
	switch level {
	case zap.DebugLevel:
		return log.DebugLevel
	case zap.InfoLevel:
		return log.InfoLevel
	case zap.WarnLevel:
		return log.WarnLevel
	case zap.ErrorLevel:
		return log.ErrorLevel
	case zap.FatalLevel:
		return log.FatalLevel
	default:
		return log.InfoLevel
	}
}
