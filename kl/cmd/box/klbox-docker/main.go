package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type config struct {
	Mounts map[string]string `json:"mounts"`
}

func main() {
	if err := Run(); err != nil {
		panic(err)
	}
}

func Run() error {
	var configFile string
	flag.StringVar(&configFile, "conf", "", "--conf /path/to/config.json")
	flag.Parse()

	if configFile == "" {
		return fmt.Errorf("no config file provided")
	}

	b, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	var c config
	err = json.Unmarshal(b, &c)
	if err != nil {
		return err
	}

	for k, v := range c.Mounts {
		if err := os.WriteFile(k, []byte(v), os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
