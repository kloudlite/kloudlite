package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

func init() {
	// Add container runtime commands
	dockerCmd.AddCommand(dockerRunCmd)
	dockerCmd.AddCommand(dockerPsCmd)
	dockerCmd.AddCommand(dockerStopCmd)
	dockerCmd.AddCommand(dockerStartCmd)
	dockerCmd.AddCommand(dockerRestartCmd)
	dockerCmd.AddCommand(dockerRmCmd)
	dockerCmd.AddCommand(dockerLogsCmd)
	dockerCmd.AddCommand(dockerExecCmd)
	dockerCmd.AddCommand(dockerKillCmd)
}

// Run command flags
var (
	runDetach      bool
	runName        string
	runEnv         []string
	runPorts       []string
	runVolumes     []string
	runWorkdir     string
	runEntrypoint  string
	runRm          bool
	runInteractive bool
	runTty         bool
)

var dockerRunCmd = &cobra.Command{
	Use:   "run [OPTIONS] IMAGE [COMMAND] [ARG...]",
	Short: "Create and run a new container",
	Example: `  kl docker run nginx
  kl docker run -d --name myapp -p 8080:80 nginx
  kl docker run -it alpine sh`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDockerRun,
}

func init() {
	dockerRunCmd.Flags().BoolVarP(&runDetach, "detach", "d", false, "Run container in background")
	dockerRunCmd.Flags().StringVar(&runName, "name", "", "Assign a name to the container")
	dockerRunCmd.Flags().StringArrayVarP(&runEnv, "env", "e", nil, "Set environment variables")
	dockerRunCmd.Flags().StringArrayVarP(&runPorts, "publish", "p", nil, "Publish a container's port(s) to the host")
	dockerRunCmd.Flags().StringArrayVarP(&runVolumes, "volume", "v", nil, "Bind mount a volume")
	dockerRunCmd.Flags().StringVarP(&runWorkdir, "workdir", "w", "", "Working directory inside the container")
	dockerRunCmd.Flags().StringVar(&runEntrypoint, "entrypoint", "", "Overwrite the default ENTRYPOINT")
	dockerRunCmd.Flags().BoolVar(&runRm, "rm", false, "Automatically remove the container when it exits")
	dockerRunCmd.Flags().BoolVarP(&runInteractive, "interactive", "i", false, "Keep STDIN open")
	dockerRunCmd.Flags().BoolVarP(&runTty, "tty", "t", false, "Allocate a pseudo-TTY")
}

func runDockerRun(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	imageName := args[0]
	var command []string
	if len(args) > 1 {
		command = args[1:]
	}

	// Parse port bindings
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for _, p := range runPorts {
		parts := strings.Split(p, ":")
		if len(parts) == 2 {
			hostPort := parts[0]
			containerPort := parts[1]
			port := nat.Port(containerPort + "/tcp")
			exposedPorts[port] = struct{}{}
			portBindings[port] = []nat.PortBinding{{HostPort: hostPort}}
		}
	}

	// Parse volume bindings
	var binds []string
	for _, v := range runVolumes {
		binds = append(binds, v)
	}

	// Build config
	config := &container.Config{
		Image:        imageName,
		Cmd:          command,
		Env:          runEnv,
		ExposedPorts: exposedPorts,
		WorkingDir:   runWorkdir,
		Tty:          runTty,
		OpenStdin:    runInteractive,
		AttachStdin:  runInteractive,
		AttachStdout: true,
		AttachStderr: true,
	}

	if runEntrypoint != "" {
		config.Entrypoint = []string{runEntrypoint}
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        binds,
		AutoRemove:   runRm,
	}

	// Create container
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, &network.NetworkingConfig{}, &ocispec.Platform{}, runName)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID

	// Start container
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	if runDetach {
		fmt.Println(containerID)
		return nil
	}

	// Attach to container
	attachResp, err := cli.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdin:  runInteractive,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	// Stream output
	if runTty {
		_, _ = io.Copy(os.Stdout, attachResp.Reader)
	} else {
		_, _ = stdcopy.StdCopy(os.Stdout, os.Stderr, attachResp.Reader)
	}

	// Wait for container to finish
	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with status %d", status.StatusCode)
		}
	}

	return nil
}

// Ps command
var (
	psAll bool
)

var dockerPsCmd = &cobra.Command{
	Use:     "ps [OPTIONS]",
	Short:   "List containers",
	Example: `  kl docker ps`,
	RunE:    runDockerPs,
}

func init() {
	dockerPsCmd.Flags().BoolVarP(&psAll, "all", "a", false, "Show all containers (default shows just running)")
}

func runDockerPs(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: psAll})
	if err != nil {
		return err
	}

	fmt.Printf("%-15s %-25s %-20s %-15s %-20s %s\n", "CONTAINER ID", "IMAGE", "COMMAND", "STATUS", "PORTS", "NAMES")
	for _, c := range containers {
		id := c.ID[:12]
		image := c.Image
		if len(image) > 25 {
			image = image[:22] + "..."
		}
		command := c.Command
		if len(command) > 20 {
			command = command[:17] + "..."
		}
		status := c.Status
		if len(status) > 15 {
			status = status[:12] + "..."
		}

		// Format ports
		var ports []string
		for _, p := range c.Ports {
			if p.PublicPort > 0 {
				ports = append(ports, fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			} else {
				ports = append(ports, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}
		portsStr := strings.Join(ports, ", ")
		if len(portsStr) > 20 {
			portsStr = portsStr[:17] + "..."
		}

		// Format names
		names := ""
		if len(c.Names) > 0 {
			names = strings.TrimPrefix(c.Names[0], "/")
		}

		fmt.Printf("%-15s %-25s %-20s %-15s %-20s %s\n", id, image, command, status, portsStr, names)
	}

	return nil
}

// Stop command
var dockerStopCmd = &cobra.Command{
	Use:     "stop CONTAINER [CONTAINER...]",
	Short:   "Stop one or more running containers",
	Example: `  kl docker stop mycontainer`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runDockerStop,
}

func runDockerStop(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, containerID := range args {
		if err := cli.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
			fmt.Fprintf(os.Stderr, "Error stopping %s: %v\n", containerID, err)
			continue
		}
		fmt.Println(containerID)
	}

	return nil
}

// Start command
var dockerStartCmd = &cobra.Command{
	Use:     "start CONTAINER [CONTAINER...]",
	Short:   "Start one or more stopped containers",
	Example: `  kl docker start mycontainer`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runDockerStart,
}

func runDockerStart(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, containerID := range args {
		if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting %s: %v\n", containerID, err)
			continue
		}
		fmt.Println(containerID)
	}

	return nil
}

// Restart command
var dockerRestartCmd = &cobra.Command{
	Use:     "restart CONTAINER [CONTAINER...]",
	Short:   "Restart one or more containers",
	Example: `  kl docker restart mycontainer`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runDockerRestart,
}

func runDockerRestart(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, containerID := range args {
		if err := cli.ContainerRestart(ctx, containerID, container.StopOptions{}); err != nil {
			fmt.Fprintf(os.Stderr, "Error restarting %s: %v\n", containerID, err)
			continue
		}
		fmt.Println(containerID)
	}

	return nil
}

// Rm command
var (
	rmForce   bool
	rmVolumes bool
)

var dockerRmCmd = &cobra.Command{
	Use:     "rm CONTAINER [CONTAINER...]",
	Short:   "Remove one or more containers",
	Example: `  kl docker rm mycontainer`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runDockerRm,
}

func init() {
	dockerRmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Force the removal of a running container")
	dockerRmCmd.Flags().BoolVarP(&rmVolumes, "volumes", "v", false, "Remove anonymous volumes associated with the container")
}

func runDockerRm(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, containerID := range args {
		if err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
			Force:         rmForce,
			RemoveVolumes: rmVolumes,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing %s: %v\n", containerID, err)
			continue
		}
		fmt.Println(containerID)
	}

	return nil
}

// Logs command
var (
	logsFollow     bool
	logsTail       string
	logsTimestamps bool
)

var dockerLogsCmd = &cobra.Command{
	Use:     "logs [OPTIONS] CONTAINER",
	Short:   "Fetch the logs of a container",
	Example: `  kl docker logs mycontainer`,
	Args:    cobra.ExactArgs(1),
	RunE:    runDockerLogs,
}

func init() {
	dockerLogsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	dockerLogsCmd.Flags().StringVar(&logsTail, "tail", "all", "Number of lines to show from the end of the logs")
	dockerLogsCmd.Flags().BoolVarP(&logsTimestamps, "timestamps", "t", false, "Show timestamps")
}

func runDockerLogs(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	containerID := args[0]

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     logsFollow,
		Tail:       logsTail,
		Timestamps: logsTimestamps,
	}

	out, err := cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return err
	}
	defer out.Close()

	_, _ = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return nil
}

// Exec command
var (
	execDetach      bool
	execInteractive bool
	execTty         bool
	execWorkdir     string
	execEnv         []string
)

var dockerExecCmd = &cobra.Command{
	Use:     "exec [OPTIONS] CONTAINER COMMAND [ARG...]",
	Short:   "Execute a command in a running container",
	Example: `  kl docker exec mycontainer ls -la
  kl docker exec -it mycontainer sh`,
	Args: cobra.MinimumNArgs(2),
	RunE: runDockerExec,
}

func init() {
	dockerExecCmd.Flags().BoolVarP(&execDetach, "detach", "d", false, "Detached mode: run command in the background")
	dockerExecCmd.Flags().BoolVarP(&execInteractive, "interactive", "i", false, "Keep STDIN open")
	dockerExecCmd.Flags().BoolVarP(&execTty, "tty", "t", false, "Allocate a pseudo-TTY")
	dockerExecCmd.Flags().StringVarP(&execWorkdir, "workdir", "w", "", "Working directory inside the container")
	dockerExecCmd.Flags().StringArrayVarP(&execEnv, "env", "e", nil, "Set environment variables")
}

func runDockerExec(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	containerID := args[0]
	command := args[1:]

	execConfig := container.ExecOptions{
		Cmd:          command,
		AttachStdin:  execInteractive,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          execTty,
		WorkingDir:   execWorkdir,
		Env:          execEnv,
	}

	execResp, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	if execDetach {
		return cli.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{Detach: true})
	}

	attachResp, err := cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Tty: execTty,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	if execTty {
		_, _ = io.Copy(os.Stdout, attachResp.Reader)
	} else {
		_, _ = stdcopy.StdCopy(os.Stdout, os.Stderr, attachResp.Reader)
	}

	return nil
}

// Kill command
var dockerKillCmd = &cobra.Command{
	Use:     "kill CONTAINER [CONTAINER...]",
	Short:   "Kill one or more running containers",
	Example: `  kl docker kill mycontainer`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runDockerKill,
}

func runDockerKill(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, containerID := range args {
		if err := cli.ContainerKill(ctx, containerID, "SIGKILL"); err != nil {
			fmt.Fprintf(os.Stderr, "Error killing %s: %v\n", containerID, err)
			continue
		}
		fmt.Println(containerID)
	}

	return nil
}
