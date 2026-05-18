package main

import (
	"os"
	"os/exec"
)

// CommandExecutor defines an interface for executing shell commands
// This allows for mocking in tests
type CommandExecutor interface {
	Execute(script string) ([]byte, error)
}

// RealCommandExecutor implements CommandExecutor using os/exec
type RealCommandExecutor struct{}

func (r *RealCommandExecutor) Execute(script string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", script)
	// Pass through parent environment variables (PATH, etc.)
	// Don't set NIX_STATE_DIR since we're mounting /nix/store and /nix/var directly
	cmd.Env = os.Environ()
	return cmd.CombinedOutput()
}

// HostCommandExecutor implements CommandExecutor using nsenter to run commands on the host
// This is necessary for GPU detection and driver installation which must happen on the host system
type HostCommandExecutor struct{}

func (r *HostCommandExecutor) Execute(script string) ([]byte, error) {
	// Use nsenter to execute commands on the host by entering the mount namespace of PID 1
	// -t 1: target PID 1 (the host's init process)
	// -m: enter mount namespace
	// -u: enter UTS namespace
	// -i: enter IPC namespace
	// Set PATH explicitly to ensure standard binaries like nvidia-smi are found
	cmd := exec.Command("nsenter", "-t", "1", "-m", "-u", "-i", "bash", "-c", script)
	cmd.Env = append(os.Environ(), "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	return cmd.CombinedOutput()
}

// FileSystem defines an interface for filesystem operations
// This allows for mocking in tests
type FileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	Chown(name string, uid, gid int) error
	WriteFile(name string, data []byte, perm os.FileMode) error
	Rename(oldpath, newpath string) error
}

// RealFileSystem implements FileSystem using os package
type RealFileSystem struct{}

func (r *RealFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (r *RealFileSystem) Chown(name string, uid, gid int) error {
	return os.Chown(name, uid, gid)
}

func (r *RealFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (r *RealFileSystem) Rename(oldpath, newpath string) error {
	// Remove target file if it exists to ensure rename succeeds
	// On some systems/mounts, os.Rename may fail if target exists
	// This maintains reasonable atomicity since remove + rename is very fast
	_ = os.Remove(newpath)
	return os.Rename(oldpath, newpath)
}
