package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type config struct {
	Mounts map[string]string `json:"mounts"`
	DNS    string            `json:"dns"`
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

		if err := os.MkdirAll(filepath.Dir(k), fs.ModePerm); err != nil {
			return err
		}

		if err := os.WriteFile(k, []byte(v), fs.ModePerm); err != nil {
			return err
		}
	}

	wgPath := "/etc/resolv.conf"
	if c.DNS != "" {
		if err := os.WriteFile(wgPath, []byte(c.DNS), fs.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
