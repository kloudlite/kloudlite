package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	v8 "rogchap.com/v8go"
)

func ValidatePrice(functionString, inputString string) (string, error) {

	f := `(` + functionString + `)(` + inputString + `)`

	ctx := v8.NewContext()                  // creates a new V8 context with a new Isolate aka VM
	val, err := ctx.RunScript(f, "math.js") // executes a script on the global context

	if err != nil {
		return "", err
	}

	o, err := json.Marshal(val)

	if err != nil {
		return "", err
	}

	return string(o), nil
}

func main() {

	app := fiber.New()
	runPort := os.Getenv("PORT")

	app.Get("/validate-price", func(c *fiber.Ctx) error {

		var data struct {
			FunctionString string
			InputString    string
		}

		err := c.BodyParser(&data)

		if err != nil {
			fmt.Println(err)
			return errors.New("You must provide: {functionstring,inputstring}")
		}

		out, err := ValidatePrice(data.FunctionString, data.InputString)

		if err != nil {
			fmt.Println(err)
			return err
		}

		return c.SendString(out)
	})

	if runPort == "" {
		runPort = "3000"
	}
	fmt.Println(app.Listen(":" + runPort))
}
