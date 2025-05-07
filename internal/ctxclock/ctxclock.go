package ctxclock

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"fknsrs.biz/p/ytmusic/internal/ctxlogger"
)

// context registation

var clockKey int

func WithClock(ctx context.Context, c Clock) context.Context {
	if c == nil {
		c = NewRealClock()
	}

	return context.WithValue(ctx, &clockKey, c)
}

func GetClock(ctx context.Context) Clock {
	if v := ctx.Value(&clockKey); v != nil {
		return v.(Clock)
	}

	return nil
}

func Now(ctx context.Context) (time.Time, error) {
	if c := GetClock(ctx); c != nil {
		return c.Now()
	}

	return time.Time{}, fmt.Errorf("ctxclock.Now: no clock source found in context")
}

// middleware

func Register(c Clock) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(WithClock(r.Context(), c)))
	}
}

func AddLoggerHooks() func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(ctxlogger.AddHookPair(
			r.Context(),
			func(rw http.ResponseWriter, r *http.Request, l logrus.FieldLogger) logrus.FieldLogger {
				if c := GetClock(r.Context()); c != nil {
					now, err := c.Now()
					if err != nil {
						l.WithError(err).Error("clock middleware could not get start time")
					} else {
						l = l.WithFields(logrus.Fields{
							"http.request_start": now.Format(time.RFC3339),
						})
					}
				} else {
					l.Warning("clock middleware could not find clock in context to get start time")
				}

				return l
			},
			func(rw http.ResponseWriter, r *http.Request, l logrus.FieldLogger) logrus.FieldLogger {
				if c := GetClock(r.Context()); c != nil {
					now, err := c.Now()
					if err != nil {
						l.WithError(err).Error("clock middleware could not get end time")
					} else {
						l = l.WithFields(logrus.Fields{
							"http.response_end": now.Format(time.RFC3339),
						})
					}
				} else {
					l.Warning("clock middleware could not find clock in context to get end time")
				}

				return l
			},
		)))
	}
}

// public interface

var (
	ErrNoTimesLeft = fmt.Errorf("ctxclock.ErrNoTimesLeft: no times left")
	ErrNoClock     = fmt.Errorf("ctxclock.ErrNoClock: no clock found in context")
)

type Clock interface {
	Now() (time.Time, error)
}

// real clock

type realClock struct{}

func NewRealClock() Clock {
	return &realClock{}
}

func (realClock) Now() (time.Time, error) {
	return time.Now(), nil
}

// static clock

type staticClock struct{ t time.Time }

func NewStaticClock(t time.Time) Clock {
	return &staticClock{t: t}
}

func (c *staticClock) Now() (time.Time, error) {
	return c.t, nil
}

// error clock

type errorClock struct{ err error }

func NewErrorClock(err error) Clock {
	return &errorClock{err: err}
}

func (c *errorClock) Now() (time.Time, error) {
	return time.Time{}, fmt.Errorf("ctxclock.errorClock.Now: %w", c.err)
}

// stacked clock, useful for testing

type stackedClock struct{ clocks []Clock }

func NewStackedClock(clocks []Clock) *stackedClock {
	return &stackedClock{clocks: clocks}
}

func (c *stackedClock) Now() (time.Time, error) {
	var errs []error

	for i, e := range c.clocks {
		t, err := e.Now()
		if err != nil {
			errs = append(errs, fmt.Errorf("ctxclock.stackedClock.Now: clock %d: %w", i, err))
		}

		return t, nil
	}

	return time.Time{}, fmt.Errorf("ctxclock.stackedClock.Now: %w", errors.Join(errs...))
}

// testing clock

type TestClockResult struct {
	Time  time.Time
	Error error
}

type testClock struct {
	m sync.RWMutex
	a []TestClockResult
	i int
}

func NewTestClock(results []TestClockResult) Clock {
	return &testClock{a: results}
}

func (c *testClock) Now() (time.Time, error) {
	c.m.RLock()
	if c.i >= len(c.a) {
		c.m.RUnlock()
		return time.Time{}, fmt.Errorf("ctxclock.testClock.Now: %w", ErrNoTimesLeft)
	}
	c.m.RUnlock()

	c.m.Lock()
	defer c.m.Unlock()

	if c.i >= len(c.a) {
		return time.Time{}, fmt.Errorf("ctxclock.testClock.Now: %w", ErrNoTimesLeft)
	}

	r := c.a[c.i]

	c.i++

	return r.Time, r.Error
}
