package types

import "fmt"

type Fstring string

func (s Fstring) Format(v ...any) string {
	return fmt.Sprintf(string(s), v...)
}

func (s Fstring) String() string {
	return string(s)
}
