package color

import "fmt"

func Color(code int) string {
	return fmt.Sprintf("\033[38;05;%dm", code)
}
func ColorReset() string {
	return "\033[0m"
}

func ColorText(text string, code int) string {
	return fmt.Sprintf("\033[38;05;%dm%s\033[0m", code, text)
}
