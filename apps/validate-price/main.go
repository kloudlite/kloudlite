package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v2"
	v8 "rogchap.com/v8go"
)

func getCompute() (string, error) {

	configPath := os.Getenv("PRICING_PATH")
	if configPath == "" {
		return "", errors.New("CAN'T FIND CONFIG")
	}

	computeObj := []map[string]any{}

	computeByte, err := os.ReadFile(configPath + "compute.yaml")
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(computeByte, &computeObj)
	if err != nil {
		return "", err
	}

	storageObj := []map[string]any{}

	storageByte, err := os.ReadFile(configPath + "storage.yaml")
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(storageByte, &storageObj)
	if err != nil {
		return "", err
	}

	lambdaObj := []map[string]any{}
	lambdaByte, err := os.ReadFile(configPath + "lambda.yaml")
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(lambdaByte, &lambdaObj)
	if err != nil {
		return "", err
	}

	ciObj := []map[string]any{}

	ciByte, err := os.ReadFile(configPath + "ci.yaml")
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(ciByte, &ciObj)
	if err != nil {
		return "", err
	}

	b := map[string]map[string]any{}

	for _, i := range ciObj {
		name := i["name"]
		b[name.(string)] = i
	}

	ci, err := json.Marshal(b)
	if err != nil {
		return "", err
	}

	b = map[string]map[string]any{}

	for _, i := range computeObj {
		name := i["name"]
		b[name.(string)] = i
	}

	compute, err := json.Marshal(b)

	if err != nil {
		return "", err
	}

	b = map[string]map[string]any{}

	for _, i := range storageObj {
		name := i["name"]
		b[name.(string)] = i
	}

	storage, err := json.Marshal(b)

	if err != nil {
		return "", err
	}

	b = map[string]map[string]any{}

	for _, i := range lambdaObj {
		name := i["name"]
		b[name.(string)] = i
	}

	lambda, err := json.Marshal(b)

	if err != nil {
		return "", err
	}

	return `{
        compute: ` + string(compute) + `,
        storage: ` + string(storage) + `,
        lambda: ` + string(lambda) + `,
        ci: ` + string(ci) + `,
      }`, nil

}

func ValidatePrice(functionString, inputString, priceDetails string) (string, error) {

	f := `(` + functionString + `)
      (` + inputString + `,
      ` + priceDetails + `
      )`

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

	priceDetails, configParseError := getCompute()

	app.Get("/validate-price", func(c *fiber.Ctx) error {
		if configParseError != nil {
			return configParseError
		}
		var data struct {
			FunctionString string
			InputString    string
		}

		err := c.BodyParser(&data)

		if err != nil {
			fmt.Println(err)
			return err
		}

		out, err := ValidatePrice(data.FunctionString, data.InputString, priceDetails)

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
