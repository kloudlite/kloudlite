/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package project

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "project",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(useCmd)
	Cmd.AddCommand(dockerCredentialsCmd)
}
