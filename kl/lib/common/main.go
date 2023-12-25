package common

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/kloudlite/kl/lib/common/ui/text"
)

type Option struct {
	Key   string
	Value string
}

func GetOption(op []Option, key string) string {
	for _, o := range op {
		if o.Key == key {
			return o.Value
		}
	}

	return ""
}

func MakeOption(key, value string) Option {
	return Option{
		Key:   key,
		Value: value,
	}
}

func PrintError(err error) {
	_, _ = os.Stderr.WriteString(fmt.Sprintf("%s\n", text.Colored(err.Error(), 1)))
}

func Log(str string) {
	_, _ = fmt.Fprintf(os.Stderr, "%s\n", str)
}

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

func GetConfigFolder() (configFolder string, err error) {
	var dirName string
	dirName, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		dirName, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	configFolder = fmt.Sprintf("%s/.kl", dirName)
	if _, err := os.Stat(configFolder); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(configFolder, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	return configFolder, nil
}
