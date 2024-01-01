package text

import "fmt"

func Color(code int) string {
	return fmt.Sprintf("\033[38;05;%dm", code)
}

func Reset() string {
	return "\033[0m"
}

func Colored(text interface{}, code int) string {
	return fmt.Sprintf("\033[38;05;%dm%v\033[0m", code, text)
}

func Bold(text interface{}) string {
	return fmt.Sprintf("\033[1m%v\033[0m", text)
}

func Underline(text interface{}) string {
	return fmt.Sprintf("\033[4m%v\033[0m", text)
}

func ColoredBold(text interface{}, code int) string {
	return fmt.Sprintf("\033[1m\033[38;05;%dm%v\033[0m", code, text)
}

func ColoredUnderline(text interface{}, code int) string {
	return fmt.Sprintf("\033[4m\033[38;05;%dm%v\033[0m", code, text)
}

func BoldUnderline(text interface{}) string {
	return fmt.Sprintf("\033[1m\033[4m%v\033[0m", text)
}

func ColoredBoldUnderline(text interface{}, code int) string {
	return fmt.Sprintf("\033[1m\033[4m\033[38;05;%dm%v\033[0m", code, text)
}

func Red(text interface{}) string {
	return Colored(text, 1)
}

func Green(text interface{}) string {
	return Colored(text, 2)
}

func Blue(text interface{}) string {
	return Colored(text, 4)
}
