package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "kloudlite",
		Short: "Kloudlite server runtime",
	}

	root.AddCommand(newServerCommand())
	return root
}

func Execute() {
	if err := NewRootCommand().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}
