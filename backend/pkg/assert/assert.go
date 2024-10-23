package assert

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
)

var red = "\033[31m"
var reset = "\033[0m"

func execAssert(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\n\n%sAssertion failed: %s%s\n", red, reset, msg)

	fmt.Fprintf(os.Stderr, "Args:\n")
	for i, v := range args {
		fmt.Fprintf(os.Stderr, "   %d: %v\n", i, v)
	}

	fmt.Fprintf(os.Stderr, "\n\n")
	fmt.Fprintln(os.Stderr, string(debug.Stack()))
	os.Exit(1)
}

func True(ok bool, msg string, data ...any) {
	if !ok {
		execAssert(msg, data...)
	}
}

func Nil(item any, msg string, data ...any) {
	if item != nil {
		execAssert(msg, data...)
	}
}

func NotNil(item any, msg string, data ...any) {
	if item == nil || reflect.ValueOf(item).Kind() == reflect.Ptr && reflect.ValueOf(item).IsNil() {
		execAssert(msg, data...)
	}
}

func Never(msg string, data ...any) {
	execAssert(msg, data...)
}
