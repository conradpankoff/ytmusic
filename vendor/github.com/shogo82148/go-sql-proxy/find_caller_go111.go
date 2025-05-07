//go:build !go1.12
// +build !go1.12

package proxy

import (
	"runtime"
	"strings"
)

func findCaller(f Filter) int {
	// skip starts 5. 0: Callers, 1: findCaller, 2: hooks, 3: proxy-funcs, 4: database/sql, and equals or greater than 5: user-funcs
	skip := 5
	for {
		var rpc [8]uintptr
		n := runtime.Callers(skip, rpc[:])

		for i, pc := range rpc[:n] {
			// http://stackoverflow.com/questions/25262754/how-to-get-name-of-current-package-in-go
			name := runtime.FuncForPC(pc).Name()
			if name == "" || strings.HasPrefix(name, "runtime.") {
				continue
			}
			dotIdx := 0
			for j := len(name) - 1; j >= 0; j-- {
				if name[j] == '.' {
					dotIdx = j
				} else if name[j] == '/' {
					break
				}
			}
			packageName := name[:dotIdx]
			if f.DoOutput(packageName) {
				return skip + i + 1
			}
		}
		if n < len(rpc) {
			break
		}
		skip += n
	}
	// fallback to the caller
	// 1: Outputter.Output, 2: the caller
	return 2
}
