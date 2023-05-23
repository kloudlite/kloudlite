package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	fn "github.com/kloudlite/helm-charts/cmd/tmpl/pkg/functions"
	"github.com/kloudlite/helm-charts/cmd/tmpl/pkg/template"
)

var AppName = "tmpl"

var Version = "v0.0.1"

func main() {
	app := cli.NewApp()
	app.Name = AppName
	app.Version = Version
	app.Commands = []*cli.Command{
		{
			Name: "parse",
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:     "set",
					Required: false,
				},
				&cli.StringFlag{
					Name:      "f",
					Usage:     "-f <path-to-template-file>",
					TakesFile: true,
					Required:  true,
				},
				&cli.StringFlag{
					Name:     "missing-key",
					Usage:    "--missing-key <path-to-template-file>",
					Value:    "default",
					Required: false,
				},
			},
			Action: func(ctx *cli.Context) error {
				setArgs := ctx.StringSlice("set")
				valueMap := make(map[string]any, (len(os.Environ())+len(setArgs))*2)

				for _, v := range os.Environ() {
					split := strings.SplitN(v, "=", 2)
					if len(split) != 2 {
						log.Printf("[WARNING]: invalid env key=value (%s)\n", v)
						continue
					}
					valueMap[split[0]] = split[1]
					valueMap[fn.Capitalize(split[0])] = split[1]
				}

				for i := range setArgs {
					split := strings.SplitN(setArgs[i], "=", 2)
					if len(split) != 2 {
						panic(fmt.Errorf("invalid set-args key=value (%s)", setArgs[i]))
					}
					valueMap[split[0]] = split[1]
					valueMap[fn.Capitalize(split[0])] = split[1]
				}

				fName := ctx.String("f")
				t := template.New(fName)

				mk := ctx.String("missing-key")
				if mk == "error" {
					t.Option("missingkey=error")
				}
				if mk == "zero" {
					t.Option("missingkey=zero")
				}

				if _, err := t.ParseFiles(fName); err != nil {
					return err
				}

				if err := t.ExecuteTemplate(os.Stdout, filepath.Base(fName), valueMap); err != nil {
					return err
				}
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
