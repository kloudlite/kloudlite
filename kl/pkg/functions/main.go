package functions

import (
	"fmt"
	"github.com/kloudlite/kl/pkg/ui/text"
	"os"
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
