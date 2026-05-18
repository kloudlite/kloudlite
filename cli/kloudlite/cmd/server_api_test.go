package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIServerCommandRunsInjectedRunner(t *testing.T) {
	called := false
	cmd := newAPIServerCommandWithRunner(func(ctx context.Context) error {
		called = true
		return nil
	})

	cmd.SetArgs([]string{})
	require.NoError(t, cmd.ExecuteContext(context.Background()))
	require.True(t, called)
}

func TestAPIServerCommandExposesInstallCRDsFlag(t *testing.T) {
	cmd := newAPIServerCommandWithRunner(func(ctx context.Context) error { return nil })
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	require.NoError(t, cmd.ExecuteContext(context.Background()))
	require.Contains(t, buf.String(), "--install-crds")
	require.Contains(t, buf.String(), "--crds-dir")
}
