package functions

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/martinlindhe/notify"
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
	_, _ = os.Stderr.WriteString(fmt.Sprintf("[#] %s\n", text.Red(err.Error())))
}

func Log(str ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, fmt.Sprint(fmt.Sprint(str...), "\n"))
}

func Warn(str ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, text.Yellow(fmt.Sprint(fmt.Sprint(str...), "\n")))
}

func Logf(format string, str ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf(fmt.Sprint(format, "\n"), str...))
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

func ParseBoolFlag(cmd *cobra.Command, flag string) bool {
	if cmd.Flags().Changed(flag) {
		v, _ := cmd.Flags().GetBool(flag)
		return v
	}

	return false
}

func WithOutputVariant(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "table", "output format [table | json | yaml]")
}

func WithKlFile(cmd *cobra.Command) {
	cmd.Flags().StringP("klfile", "k", "", "kloudlite file")
}

func ParseKlFile(cmd *cobra.Command) string {
	if cmd.Flags().Changed("klfile") {
		v, _ := cmd.Flags().GetString("klfile")
		return v
	}

	return ""
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

func Alert(name string, str ...interface{}) {
	if runtime.GOOS == "darwin" {
		notify.Alert("Kloudlite", name, fmt.Sprint(str...), "")
	}
	if runtime.GOOS == "linux" {
		notification(name, fmt.Sprint(str...), "")
		if err := exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/alarm-clock-elapsed.oga").Start(); err != nil {
			log.Println("error playing alert sound:", err)
		}
	}
}

func Notify(name string, str ...interface{}) {
	if runtime.GOOS == "darwin" {
		notify.Notify("Kloudlite", name, fmt.Sprint(str...), "")
	}

	if runtime.GOOS == "linux" {
		notification(name, fmt.Sprint(str...), "")
	}
}

func Desc(str string) string {
	return strings.Replace(str, "{cmd}", flags.CliName, -1)
}

func notification(title string, txt string, iconPath string) {
	if euid := os.Geteuid(); euid == 0 {
		if usr, ok := os.LookupEnv("SUDO_USER"); ok {
			if euid, ok := os.LookupEnv("SUDO_UID"); ok {
				c := fmt.Sprintf("sudo -u %s DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/%s/bus notify-send -i %q %q %q", usr, euid, iconPath, title, txt)
				if err := ExecCmd(c, nil, false); err != nil {
					PrintError(err)
				}
			}
		}

		return
	}

	if err := ExecCmd(fmt.Sprintf("notify-send -i %q %q %q", iconPath, title, txt), nil, false); err != nil {
		PrintError(err)
	}
}

func Truncate(str string, length int) string {
	if len(str) < length {
		return str
	}

	return fmt.Sprintf("%s...", str[0:length])
}
