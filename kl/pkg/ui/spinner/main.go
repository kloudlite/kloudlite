package spinner

import (
	"fmt"
	"github.com/briandowns/spinner"
	"time"
)

func NewSpinner(msg ...string) *spinner.Spinner {
	var message []string
	sp := spinner.CharSets[11]
	for _, v := range sp {
		message = append(message, fmt.Sprintf("%s %s...", v, func() string {
			if len(msg) == 0 {
				return "loading please wait"
			}
			return msg[0]
		}()))
	}
	return spinner.New(message, 100*time.Millisecond)
}
