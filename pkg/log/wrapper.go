package log

import (
	"context"
	"os"

	"github.com/ppxb/oreo-admin-go/pkg/constant"
	"github.com/ppxb/oreo-admin-go/pkg/tracing"
)

type Wrapper struct {
	log    Interface
	fields map[string]interface{}
}

func NewWrapper(l Interface) *Wrapper {
	return &Wrapper{
		log:    l,
		fields: map[string]interface{}{},
	}
}

func (w *Wrapper) logWithLevel(level Level, args ...interface{}) {
	if !w.log.Options().level.Enabled(level) {
		return
	}
	ns := w.prepareFields()
	if len(args) > 1 {
		if format, ok := args[0].(string); ok {
			w.log.WithFields(ns).Logf(level, format, args[1:]...)
			return
		}
	}
	w.log.WithFields(ns).Log(level, args...)
}

func (w *Wrapper) prepareFields() map[string]interface{} {
	ns := copyFields(w.fields)
	if w.log.Options().lineNum {
		ns[constant.LogLineNumKey] = fileWithLineNum(w.log.Options())
	}
	return ns
}

func (w *Wrapper) Trace(args ...interface{}) {
	w.logWithLevel(TraceLevel, args...)
}

func (w *Wrapper) Debug(args ...interface{}) {
	w.logWithLevel(DebugLevel, args...)
}

func (w *Wrapper) Info(args ...interface{}) {
	w.logWithLevel(InfoLevel, args...)
}

func (w *Wrapper) Warn(args ...interface{}) {
	w.logWithLevel(WarnLevel, args...)
}

func (w *Wrapper) Error(args ...interface{}) {
	w.logWithLevel(ErrorLevel, args...)
}

func (w *Wrapper) Fatal(args ...interface{}) {
	w.logWithLevel(FatalLevel, args...)
	os.Exit(1)
}

func (w *Wrapper) WithError(err error) *Wrapper {
	ns := copyFields(w.fields)
	ns[constant.LogErrorKey] = err
	return &Wrapper{
		log:    w.log,
		fields: ns,
	}
}

func (w *Wrapper) WithFields(fields map[string]interface{}) *Wrapper {
	ns := copyFields(fields)
	for k, v := range w.fields {
		ns[k] = v
	}
	return &Wrapper{
		log:    w.log,
		fields: ns,
	}
}

func (w *Wrapper) WithContext(ctx context.Context) *Wrapper {
	requestId, traceId, spanId := tracing.GetId(ctx)
	if requestId == "" {
		return w
	}
	ns := copyFields(w.fields)
	if traceId != "" {
		ns[constant.MiddlewareTraceIdCtxKey] = traceId
		ns[constant.MiddlewareSpanIdCtxKey] = spanId
	} else {
		ns[constant.MiddlewareRequestIdCtxKey] = requestId
	}
	return &Wrapper{
		log:    w.log,
		fields: ns,
	}
}

func copyFields(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
