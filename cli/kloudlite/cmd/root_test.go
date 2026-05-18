package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRootCommandIncludesServerModes(t *testing.T) {
	root := NewRootCommand()
	require.Equal(t, "kloudlite", root.Use)

	serverCmd, _, err := root.Find([]string{"server"})
	require.NoError(t, err)
	require.Equal(t, "server", serverCmd.Use)

	apiCmd, _, err := root.Find([]string{"server", "api"})
	require.NoError(t, err)
	require.Equal(t, "api", apiCmd.Use)

	tunnelCmd, _, err := root.Find([]string{"server", "tunnel"})
	require.NoError(t, err)
	require.Equal(t, "tunnel", tunnelCmd.Use)
}
