package fileclient

import (
	"github.com/spf13/cobra"
)

type fclient struct {
}

type FileClient interface {
	WriteConfigFile(fileObj ConfigFile) error
	GetConfigFile() (*ConfigFile, error)
}

func New(cmd *cobra.Command) (FileClient, error) {
	return &fclient{}, nil
}
