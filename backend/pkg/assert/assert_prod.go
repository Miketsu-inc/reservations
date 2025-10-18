//go:build prod

package assert

func True(ok bool, msg string, data ...any)    {}
func Nil(item any, msg string, data ...any)    {}
func NotNil(item any, msg string, data ...any) {}
func Never(msg string, data ...any)            {}
