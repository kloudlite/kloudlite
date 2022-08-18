package common

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
}

func NewSpinner() *spinner.Spinner {
	message := []string{}
	sp := spinner.CharSets[11]
	for _, v := range sp {
		message = append(message, fmt.Sprintf("%s loading please wait...", v))
	}
	return spinner.New(message, 100*time.Millisecond)
}
