/*
Copyright © 2022 Kloudlite <support@kloudlite.io>

*/
package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type ResEnvType struct {
	Key    string
	RefKey string
}

type EnvType struct {
	Key   string
	Value string
}

type ResType struct {
	Id   string
	Name string
	Env  []ResEnvType
}

type KLFileType struct {
	Version string
	Name    string
	Mres    []ResType
	Configs []ResType
	Secrets []ResType
	Env     []EnvType
	Ports   []string
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		initFile := KLFileType{
			Version: "v1",
			Name:    "Sample",
			Mres:    []ResType{},
			Configs: []ResType{},
			Secrets: []ResType{},
			Env:     []EnvType{},
			Ports:   []string{},
		}

		file, err := yaml.Marshal(initFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = ioutil.WriteFile(".kl.yml", file, 0644)

		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
