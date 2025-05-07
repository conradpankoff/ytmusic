package ctxtimer

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"fknsrs.biz/p/ytmusic/internal/ctxclock"
	"fknsrs.biz/p/ytmusic/internal/ctxlogger"
)

// context registration

var timerKey int

func WithTimer(ctx context.Context, t Timer) context.Context {
	if t == nil {
		t = NewTimer()
	}

	return context.WithValue(ctx, &timerKey, t)
}

func GetTimer(ctx context.Context) Timer {
	if v := ctx.Value(&timerKey); v != nil {
		t := v.(Timer)

		if c := ctxclock.GetClock(ctx); c != nil {
			return &timerWithClock{timer: t, clock: c}
		}

		return t
	}

	return nil
}

// middleware

const (
	timerNameOuter = "ctxtimer.middleware"
)

func Register(c Timer) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(WithTimer(r.Context(), c)))
	}
}

func AddLoggerHooks() func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(ctxlogger.AddHookPair(
			r.Context(),
			func(rw http.ResponseWriter, r *http.Request, l logrus.FieldLogger) logrus.FieldLogger {
				if t := GetTimer(r.Context()); t != nil {
					if t, ok := t.(TimerNow); ok {
						if err := t.MarkNow(timerNameOuter); err != nil {
							l.WithError(err).Error("ctxtimer: could not mark request start with internal clock")
						}
					} else {
						l.Warning("ctxtimer: could not find timer with knowledge of current time to mark request start")
					}
				} else {
					l.Warning("ctxtimer: could not find timer in context to mark request start")
				}

				return l
			},
			func(rw http.ResponseWriter, r *http.Request, l logrus.FieldLogger) logrus.FieldLogger {
				if t := GetTimer(r.Context()); t != nil {
					if t, ok := t.(TimerNow); ok {
						elapsed, err := t.ElapsedNow(timerNameOuter)
						if err != nil {
							l.WithError(err).Error("ctxtimer: could not get elapsed time with internal clock")
						} else {
							l = l.WithFields(logrus.Fields{"http.duration": elapsed})
						}
					} else {
						l.Warning("ctxtimer: could not find timer with knowledge of current time to get elapsed time")
					}
				} else {
					l.Warning("ctxtimer: could not find timer in context to get elapsed time")
				}

				return l
			},
		)))
	}
}

// public interface

var (
	ErrNoTimer = fmt.Errorf("ctxtimer.ErrNoTimer: no timer found with this name")
)

type Timer interface {
	Mark(name string, t time.Time)
	Elapsed(name string, t time.Time) (time.Duration, error)
}

type TimerNow interface {
	Timer
	MarkNow(name string) error
	ElapsedNow(name string) (time.Duration, error)
}

// basic timer registration

type timer struct {
	rw    sync.RWMutex
	start map[string]time.Time
}

func NewTimer() Timer {
	return &timer{start: make(map[string]time.Time)}
}

func (t *timer) Mark(name string, tt time.Time) {
	t.rw.Lock()
	defer t.rw.Unlock()

	t.start[name] = tt
}

func (t *timer) Elapsed(name string, tt time.Time) (time.Duration, error) {
	t.rw.RLock()
	defer t.rw.RUnlock()

	start, ok := t.start[name]
	if !ok {
		return 0, ErrNoTimer
	}

	return tt.Sub(start), nil
}

// timer with knowledge of current time

type timerWithClock struct {
	timer Timer
	clock ctxclock.Clock
}

func (t *timerWithClock) Mark(name string, tt time.Time) {
	t.timer.Mark(name, tt)
}

func (t *timerWithClock) Elapsed(name string, tt time.Time) (time.Duration, error) {
	d, err := t.timer.Elapsed(name, tt)
	if err != nil {
		return 0, fmt.Errorf("ctxtimer.timerWithClock.Elapsed: %w", err)
	}

	return d, nil
}

func (t *timerWithClock) MarkNow(name string) error {
	now, err := t.clock.Now()
	if err != nil {
		return fmt.Errorf("ctxtimer.timerWithClock.MarkNow: %w", err)
	}

	t.Mark(name, now)

	return nil
}

func (t *timerWithClock) ElapsedNow(name string) (time.Duration, error) {
	now, err := t.clock.Now()
	if err != nil {
		return 0, fmt.Errorf("ctxtimer.timerWithClock.ElapsedNow: %w", err)
	}

	d, err := t.Elapsed(name, now)
	if err != nil {
		return 0, fmt.Errorf("ctxtimer.timerWithClock.ElapsedNow: %w", err)
	}

	return d, nil
}
