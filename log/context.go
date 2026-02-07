package log

import "context"

type loggerKey struct{}
type wrapLoggerKey struct{}

func FromContext(ctx context.Context) (Logger, bool) {
	l, ok := ctx.Value(loggerKey{}).(Logger)
	return l, ok
}

func NewContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// 将 l 注入到 ctx 中
func Inject_RequestID(ctx context.Context, requestID string) context.Context {

	wl := NewHelper(DefaultLogger).WithFields(map[string]interface{}{"requestID": requestID})

	return context.WithValue(ctx, wrapLoggerKey{}, wl)
}

// 从ctx中，拿到log （函数名为什么取L，也就是为了短）
func L(ctx context.Context) *Helper {
	if l, ok := ctx.Value(wrapLoggerKey{}).(*Helper); ok {
		return l
	}

	return NewHelper(DefaultLogger)
}
