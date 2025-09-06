package helper

import (
	"runtime"
	"strings"
)

func ThisFuncName() string {
	// Get the function name using reflection
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	// Extract just the function name part (after last dot)
	i := strings.LastIndex(funcName, ".")
	if i < 0 || i+1 > len(funcName) {
		return funcName
	}
	fn := funcName[i+1:]
	if fn == "" {
		return funcName
	}
	return fn
}
