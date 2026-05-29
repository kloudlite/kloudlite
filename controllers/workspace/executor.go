package workspace

import (
	"os"
	"os/exec"
)

// CommandExecutor defines an interface for executing shell commands.
// This allows for mocking in tests and using nsenter for host-level
// commands like btrfs subvolume management.
type CommandExecutor interface {
	Execute(script string) ([]byte, error)
}

// HostCommandExecutor implements CommandExecutor using nsenter to run commands
// on the host. This is necessary for btrfs subvolume management which must
// happen on the host filesystem.
type HostCommandExecutor struct{}

func (r *HostCommandExecutor) Execute(script string) ([]byte, error) {
	cmd := exec.Command("nsenter", "-t", "1", "-m", "-u", "-i", "bash", "-c", script)
	cmd.Env = append(os.Environ(), "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	return cmd.CombinedOutput()
}
