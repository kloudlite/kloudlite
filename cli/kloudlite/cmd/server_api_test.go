package cmd

import (
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
