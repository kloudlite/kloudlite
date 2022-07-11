package main

import "fmt"

type A struct {
	name  string
	value string
}

func main() {
	fmt.Println(A{name: "name", value: "value"} == A{name: "name", value: "value"})
}
