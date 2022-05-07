package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var setupCommand = &cobra.Command{
	Use:   "setup",
	Short: "Setup Cluster",
	Long:  `This will setup new shared cluster`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var rootCmd = &cobra.Command{
	Use:   "kluster",
	Short: "kluster",
	Long:  `kloudlite client`,
}

func init() {
	rootCmd.AddCommand(setupCommand)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
