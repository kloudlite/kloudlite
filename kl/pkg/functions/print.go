package functions

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func stderr(str string) {
	if spinner.Client.IsRunning() {
		spinner.Client.Pause()
		defer spinner.Client.Resume()
	}

	_, _ = os.Stderr.WriteString(str)
}

func stdout(str string) {
	if spinner.Client.IsRunning() {
		spinner.Client.Pause()
		defer spinner.Client.Resume()
	}

	_, _ = os.Stdout.WriteString(str)
}

func Log(str ...interface{}) {
	stderr(fmt.Sprint(fmt.Sprint(str...), "\n"))
}

func Warn(str ...interface{}) {
	if flags.IsQuiet {
		return
	}

	stderr(fmt.Sprintf("%s %s", text.Yellow("[warn]"), fmt.Sprint(fmt.Sprint(str...), "\n")))
}

func Warnf(format string, str ...interface{}) {
	if flags.IsQuiet {
		return
	}

	stderr(fmt.Sprintf(format, str...))
}

func Logf(format string, str ...interface{}) {
	if flags.IsQuiet {
		return
	}

	stderr(fmt.Sprintf(format, str...))
}

func Printf(format string, str ...interface{}) {
	if flags.IsQuiet {
		return
	}

	stdout(fmt.Sprintf(format, str...))
}

func Println(str ...interface{}) {
	if flags.IsQuiet {
		return
	}

	stdout(fmt.Sprint(str...))
}
