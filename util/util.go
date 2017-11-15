package util

import (
	"reflect"
	"runtime"
)

// FunctionName returns the name of the function for easier debugging.
func FunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
