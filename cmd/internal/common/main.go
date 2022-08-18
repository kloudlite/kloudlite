package common

import (
	"fmt"
	"os"
)

func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
}
