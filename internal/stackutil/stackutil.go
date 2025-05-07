package stackutil

import (
	"fmt"
	"runtime"
)

func GetStack(depth, skip int) []runtime.Frame {
	pc := make([]uintptr, depth)

	n := runtime.Callers(0, pc)
	if n == 0 {
		return []runtime.Frame{}
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	// skip this call and the one to runtime.Callers
	skip += 2

	var a []runtime.Frame

	for i := 0; ; i++ {
		frame, more := frames.Next()

		if i >= skip {
			a = append(a, frame)
		}

		if !more {
			break
		}
	}

	return a
}

func FormatStack(a []runtime.Frame) []string {
	r := make([]string, len(a))
	for i, e := range a {
		r[i] = FormatStackFrame(e)
	}
	return r
}

func FormatStackFrame(f runtime.Frame) string {
	return fmt.Sprintf("%s:%d: %s", f.File, f.Line, f.Function)
}
