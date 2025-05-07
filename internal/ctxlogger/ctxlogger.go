package ctxlogger

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

// context registration

var loggerKey int

func WithLogger(ctx context.Context, l logrus.FieldLogger) context.Context {
	return context.WithValue(ctx, &loggerKey, l)
}

func GetLogger(ctx context.Context) logrus.FieldLogger {
	if v := ctx.Value(&loggerKey); v != nil {
		return v.(logrus.FieldLogger)
	}

	return logrus.StandardLogger()
}

// middleware

var (
	hookListKey int
)

type Hook interface {
	Before(rw http.ResponseWriter, r *http.Request, l logrus.FieldLogger) logrus.FieldLogger
	After(rw http.ResponseWriter, r *http.Request, l logrus.FieldLogger) logrus.FieldLogger
}

type HookFunc func(rw http.ResponseWriter, r *http.Request, l logrus.FieldLogger) logrus.FieldLogger

type HookPair struct {
	BeforeFunc HookFunc
	AfterFunc  HookFunc
}

func NewHookPair(beforeFunc, afterFunc HookFunc) *HookPair {
	return &HookPair{BeforeFunc: beforeFunc, AfterFunc: afterFunc}
}

func (p *HookPair) Before(rw http.ResponseWriter, r *http.Request, l logrus.FieldLogger) logrus.FieldLogger {
	if p.BeforeFunc == nil {
		return l
	}

	return p.BeforeFunc(rw, r, l)
}

func (p *HookPair) After(rw http.ResponseWriter, r *http.Request, l logrus.FieldLogger) logrus.FieldLogger {
	if p.AfterFunc == nil {
		return l
	}

	return p.AfterFunc(rw, r, l)
}

type hookList struct {
	a []Hook
}

func (h *hookList) addHook(e Hook) {
	h.a = append(h.a, e)
}

func (h *hookList) runBefore(rw http.ResponseWriter, r *http.Request, l logrus.FieldLogger) logrus.FieldLogger {
	for _, hook := range h.a {
		l = hook.Before(rw, r, l)
	}

	return l
}

func (h *hookList) runAfter(rw http.ResponseWriter, r *http.Request, l logrus.FieldLogger) logrus.FieldLogger {
	for _, hook := range h.a {
		l = hook.After(rw, r, l)
	}

	return l
}

func withHookList(ctx context.Context, hooks *hookList) context.Context {
	return context.WithValue(ctx, &hookListKey, hooks)
}

func getHookList(ctx context.Context) *hookList {
	if v := ctx.Value(&hookListKey); v != nil {
		return v.(*hookList)
	}

	return nil
}

func ensureHookList(ctx context.Context) (*hookList, context.Context) {
	if hooks := getHookList(ctx); hooks != nil {
		return hooks, ctx
	}

	hooks := &hookList{}

	return hooks, withHookList(ctx, hooks)
}

func AddHook(ctx context.Context, hook Hook) context.Context {
	hooks, ctx := ensureHookList(ctx)
	hooks.addHook(hook)
	return ctx
}

func AddHookPair(ctx context.Context, beforeFunc, afterFunc HookFunc) context.Context {
	return AddHook(ctx, NewHookPair(beforeFunc, afterFunc))
}

func Register(l logrus.FieldLogger) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(WithLogger(withHookList(r.Context(), &hookList{}), l)))
	}
}

func Log() func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		hooks := getHookList(r.Context())

		l := GetLogger(r.Context())

		l = l.WithFields(logrus.Fields{
			"http.method":     r.Method,
			"http.path":       r.URL.String(),
			"http.host":       r.Host,
			"http.referer":    r.Header.Get("referer"),
			"http.user_agent": r.Header.Get("user-agent"),
		})

		if hooks != nil {
			l = hooks.runBefore(rw, r, l)
		}

		defer func() {
			if nrw, ok := rw.(interface {
				Status() int
				Size() int
			}); ok {
				l = l.WithFields(logrus.Fields{
					"http.status_code":   nrw.Status(),
					"http.response_size": nrw.Size(),
				})
			}

			if hooks != nil {
				l = hooks.runAfter(rw, r, l)
			}

			l.Info("http request finished")
		}()

		l.Info("http request started")

		next(rw, r)
	}
}
