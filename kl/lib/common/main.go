package common

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"
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
