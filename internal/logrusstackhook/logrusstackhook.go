package logrusstackhook

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"

	"fknsrs.biz/p/ytmusic/internal/stackutil"
)

type Formatter interface {
	FormatKey(i int, frame runtime.Frame) (string, error)
	FormatValue(i int, frame runtime.Frame) (string, error)
}

type FastFormatter struct{}

func (FastFormatter) FormatKey(i int, frame runtime.Frame) (string, error) {
	return fmt.Sprintf("stack.%02d", i), nil
}

func (FastFormatter) FormatValue(i int, frame runtime.Frame) (string, error) {
	return stackutil.FormatStackFrame(frame), nil
}

type FilterFunc func(index int, frame runtime.Frame) bool

func RemoveFirst(count int) FilterFunc {
	return func(index int, frame runtime.Frame) bool {
		if index < count {
			return false
		}

		return true
	}
}

func RemovePathsContaining(values []string) FilterFunc {
	return func(index int, frame runtime.Frame) bool {
		for _, value := range values {
			if strings.Contains(frame.File, value) {
				return false
			}
		}

		return true
	}
}

func RemoveFunctionsContaining(values []string) FilterFunc {
	return func(index int, frame runtime.Frame) bool {
		for _, value := range values {
			if strings.Contains(frame.Function, value) {
				return false
			}
		}

		return true
	}
}

func CombineFilters(a ...FilterFunc) FilterFunc {
	return func(index int, frame runtime.Frame) bool {
		for _, fn := range a {
			if !fn(index, frame) {
				return false
			}
		}

		return true
	}
}

var (
	DefaultFormatter = &FastFormatter{}
	DefaultLevels    = []logrus.Level{logrus.DebugLevel, logrus.TraceLevel}
	DefaultFilter    = CombineFilters(RemovePathsContaining([]string{"github.com/sirupsen/logrus"}))
)

type StackHook struct {
	formatter Formatter
	levels    []logrus.Level
	filter    FilterFunc
}

func NewStackHook(formatter Formatter, levels []logrus.Level, filter FilterFunc) *StackHook {
	if formatter == nil {
		formatter = DefaultFormatter
	}

	if levels == nil {
		levels = DefaultLevels
	}

	if filter == nil {
		filter = DefaultFilter
	}

	return &StackHook{
		formatter: formatter,
		levels:    levels,
		filter:    filter,
	}
}

func NewStackHookWithDefaults() *StackHook {
	return NewStackHook(nil, nil, nil)
}

func (h *StackHook) SetFormatter(formatter Formatter) {
	h.formatter = formatter
}

func (h *StackHook) SetLevels(levels []logrus.Level) {
	h.levels = levels
}

func AllLevelsAbove(lowestLevel logrus.Level) []logrus.Level {
	var levels []logrus.Level

	for _, e := range logrus.AllLevels {
		if e >= lowestLevel {
			levels = append(levels, e)
		}
	}

	return levels
}

func (h *StackHook) Levels() []logrus.Level { return h.levels }

func (h *StackHook) Fire(e *logrus.Entry) error {
	// skip calls from inside logrus
	for index, frame := range stackutil.GetStack(25, 0) {
		if h.filter != nil {
			if h.filter(index, frame) == false {
				continue
			}
		}

		key, err := h.formatter.FormatKey(index, frame)
		if err != nil {
			return err
		}

		value, err := h.formatter.FormatValue(index, frame)
		if err != nil {
			return err
		}

		e.Data[key] = value
	}

	return nil
}
