package helpers

import (
	"runtime"
)

const (
	defaultDepth = 1
	defaultIndex = 0
)

// GetFunctionName returns packageName.functionName
func GetFunctionName(depthList ...int) string {
	var depth int

	if depthList == nil {
		depth = defaultDepth
	} else {
		depth = depthList[defaultIndex]
	}

	function, _, _, ok := runtime.Caller(depth)
	if !ok {
		return "failed to get function name"
	}

	return runtime.FuncForPC(function).Name()
}
