package functions

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func stderr(str string) {
	if flags.IsQuiet {
		return
	}

	if spinner.Client.IsRunning() {
		spinner.Client.Pause()
		defer spinner.Client.Resume()
	}

	_, _ = os.Stderr.WriteString(str)
}

func stdout(str string) {
	if flags.IsQuiet {
		return
	}

	if spinner.Client.IsRunning() {
		spinner.Client.Pause()
		defer spinner.Client.Resume()
	}

	_, _ = os.Stdout.WriteString(str)
}

func Log(str ...interface{}) {
	str = append(str, "\n")
	stderr(fmt.Sprint(str...))
}

func Warn(str ...interface{}) {
	str = append(str, "\n")
	stderr(fmt.Sprintf("%s %s", text.Yellow("[warn]"), fmt.Sprint(str...)))
}

func Warnf(format string, str ...interface{}) {
	if spinner.Client.IsRunning() {
		spinner.Client.Pause()
	}
	stderr(fmt.Sprintf(format, str...))
}

func Logf(format string, str ...interface{}) {
	if spinner.Client.IsRunning() {
		spinner.Client.Pause()
	}
	stderr(fmt.Sprintf(format, str...))
}

func Printf(format string, str ...interface{}) {
	if spinner.Client.IsRunning() {
		spinner.Client.Pause()
	}
	stdout(fmt.Sprintf(format, str...))
}

func Println(str ...interface{}) {
	str = append(str, "\n")
	stdout(fmt.Sprint(str...))
}
