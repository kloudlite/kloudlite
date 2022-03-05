package main

import (
	"fmt"
	"kloudlite.io/pkg/errors"
)

func main() {
	const op errors.Op = "main"
	fmt.Println("hello world from messaging")
}
