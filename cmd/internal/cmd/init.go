/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"log"
	"os"
)

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
		if _, err := os.Stat("/path/to/whatever"); errors.Is(err, os.ErrNotExist) {
			fmt.Println("This will initialize a new kloudlite application. Please select a project name:")
			type Track struct {
				Name      string
				AlbumName string
				Artist    string
			}
			var tracks = []Track{
				{"foo", "album1", "artist1"},
				{"bar", "album1", "artist1"},
				{"foo", "album2", "artist1"},
				{"baz", "album2", "artist2"},
				{"baz", "album3", "artist2"},
			}

			idx, err := fuzzyfinder.Find(
				tracks,
				func(i int) string {
					return tracks[i].Name
				},
				fuzzyfinder.WithPromptString("Project:"),
			)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("selected: %v\n", idx)
			return
		}
		fmt.Println("Application is already initialized")
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
