package functions

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
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

func Log(str ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, fmt.Sprint(str...))
}

func Printf(format string, str ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format, str...)
}

func Println(str ...interface{}) {
	_, _ = fmt.Println(str...)
}

type resType struct {
	Metadata struct {
		Name string
	}
}

func GetPrintRow2(printValue interface{}, active bool, prefix ...bool) string {
	if active {
		return text.Green(fmt.Sprintf("%s%s",
			func() string {
				if len(prefix) > 0 && prefix[0] {
					return "*"
				}

				return ""
			}(),

			func() string {
				s := strings.Split(fmt.Sprint(printValue), "\n")
				if len(s) > 1 {
					for i, v := range s {
						s[i] = text.Green(v)
					}
				}

				return strings.Join(s, "\n")
			}(),
		))
	}

	return fmt.Sprint(printValue)
}

func GetPrintRow(res any, activeName string, printValue interface{}, prefix ...bool) string {
	var item resType
	if err := JsonConversion(res, &item); err != nil {
		return fmt.Sprint(printValue)
	}

	if item.Metadata.Name == activeName {
		return text.Green(fmt.Sprintf("%s%s",
			func() string {
				if len(prefix) > 0 && prefix[0] {
					return "*"
				}

				return ""
			}(),

			func() string {
				s := strings.Split(fmt.Sprint(printValue), "\n")
				if len(s) > 1 {
					for i, v := range s {
						s[i] = text.Green(v)
					}
				}

				return strings.Join(s, "\n")
			}(),
		))
	}

	return fmt.Sprint(printValue)
}

func JsonConversion(from any, to any) error {
	if from == nil {
		return nil
	}

	if to == nil {
		return fmt.Errorf("receiver (to) is nil")
	}

	b, err := json.Marshal(from)
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(b, &to); err != nil {
		return err
	}
	return nil
}

func ParseStringFlag(cmd *cobra.Command, flag string) string {
	if cmd.Flags().Changed(flag) {
		v, _ := cmd.Flags().GetString(flag)
		return v
	}

	return ""
}

func WithOutputVariant(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "table", "output format [table | json | yaml]")
}

func InfraMarkOption() Option {
	return MakeOption("isInfra", "yes")
}

func IsInfraFlagAvailable(options ...Option) bool {
	s := GetOption(options, "isInfra")
	if s == "yes" {
		return true
	}
	return false
}
