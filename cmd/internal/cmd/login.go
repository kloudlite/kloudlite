/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/lib/server"
	"os/exec"
)

const loginUrl = "https://auth.local.kl.madhouselabs.io/cli-login"

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		loginId, err := server.CreateRemoteLogin()
		if err != nil {
			fmt.Println(err)
			return
		}
		command := exec.Command("open", fmt.Sprintf("%s/%s%s", loginUrl, "?loginId=", loginId))
		err = command.Run()
		if err != nil {
			fmt.Println(err)
			return
		}
		err = server.Login(loginId)
		if err != nil {
			fmt.Println(err)
			return
		}
		TriggerSelectAccount()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
