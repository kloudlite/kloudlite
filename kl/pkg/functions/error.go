package functions

import (
	"fmt"

	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/ztrue/tracerr"
)

var wraperr = tracerr.Wrap

func NewE(err error, s ...string) error {
	if len(s) == 0 {
		return wraperr(err)
	}

	return wraperr(fmt.Errorf("%s as [%w]", s[0], err))
}

func Error(s string) error {
	return wraperr(fmt.Errorf(s))
}

func Errorf(format string, args ...interface{}) error {
	return wraperr(fmt.Errorf(format, args...))
}

// func PrintError(err error) {
// 	if err == nil {
// 		return
// 	}
//
// 	type stackTracer interface {
// 		StackTrace() errors.StackTrace
// 	}
//
// 	if flags.IsDev() {
// 		_, _ = os.Stderr.WriteString(fmt.Sprintf("%s %s\n\n", text.Red("[error]"), err.Error()))
//
// 		if err, ok := err.(stackTracer); ok {
// 			st := err.StackTrace()
// 			Logf("%s%+v", text.Bold("Stack Trace:"), st)
// 			return
// 		}
//
// 		Logf("%s\n%+v\n", text.Bold("Stack Trace:"), err)
// 		return
// 	}
//
// 	Logf("%s %s\n\n", text.Red("[error]"), err.Error())
// }

func PrintError(err error) {
	if err == nil {
		return
	}

	if flags.IsDev() || flags.IsVerbose {
		tracerr.Print(err)
		return
	}

	Logf("%s %s\n\n", text.Red("[error]"), err.Error())
}
