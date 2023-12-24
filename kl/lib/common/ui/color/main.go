package color

import "fmt"

func Color(code int) string {
	return fmt.Sprintf("\033[38;05;%dm", code)
}
func Reset() string {
	return "\033[0m"
}

func Text(text interface{}, code int) string {
	return fmt.Sprintf("\033[38;05;%dm%v\033[0m", code, text)
}
