package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker commands for building and managing container images",
	Long: `Docker commands for building and managing container images.

This command provides Docker functionality for image management operations.
Container runtime commands (run, exec, start, etc.) are not supported in workspaces.

The DOCKER_HOST environment variable is automatically configured to connect
to the Docker daemon running in your workspace namespace.`,
}

func init() {
	RootCmd.AddCommand(dockerCmd)

	// Add subcommands
	dockerCmd.AddCommand(dockerBuildCmd)
	dockerCmd.AddCommand(dockerPushCmd)
	dockerCmd.AddCommand(dockerPullCmd)
	dockerCmd.AddCommand(dockerTagCmd)
	dockerCmd.AddCommand(dockerImagesCmd)
	dockerCmd.AddCommand(dockerRmiCmd)
	dockerCmd.AddCommand(dockerInspectCmd)
	dockerCmd.AddCommand(dockerLoginCmd)
	dockerCmd.AddCommand(dockerLogoutCmd)
	dockerCmd.AddCommand(dockerInfoCmd)
	dockerCmd.AddCommand(dockerVersionCmd)
}

// getDockerClient creates a Docker client using DOCKER_HOST env var
func getDockerClient() (*client.Client, error) {
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		return nil, fmt.Errorf("DOCKER_HOST environment variable not set. Are you running inside a Kloudlite workspace?")
	}

	return client.NewClientWithOpts(
		client.WithHost(dockerHost),
		client.WithAPIVersionNegotiation(),
	)
}

// Build command
var (
	buildTags      []string
	buildFile      string
	buildNoCache   bool
	buildBuildArgs []string
	buildTarget    string
	buildPush      bool
)

var dockerBuildCmd = &cobra.Command{
	Use:   "build [OPTIONS] PATH",
	Short: "Build an image from a Dockerfile",
	Example: `  kl docker build -t myapp:latest .
  kl docker build -t myapp:latest -f Dockerfile.prod .
  kl docker build -t myapp:latest --build-arg VERSION=1.0 .`,
	Args: cobra.ExactArgs(1),
	RunE: runDockerBuild,
}

func init() {
	dockerBuildCmd.Flags().StringArrayVarP(&buildTags, "tag", "t", nil, "Name and optionally a tag (format: name:tag)")
	dockerBuildCmd.Flags().StringVarP(&buildFile, "file", "f", "Dockerfile", "Name of the Dockerfile")
	dockerBuildCmd.Flags().BoolVar(&buildNoCache, "no-cache", false, "Do not use cache when building")
	dockerBuildCmd.Flags().StringArrayVar(&buildBuildArgs, "build-arg", nil, "Set build-time variables")
	dockerBuildCmd.Flags().StringVar(&buildTarget, "target", "", "Set the target build stage")
	dockerBuildCmd.Flags().BoolVar(&buildPush, "push", false, "Push image after build")
}

func runDockerBuild(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	contextPath := args[0]
	absContextPath, err := filepath.Abs(contextPath)
	if err != nil {
		return fmt.Errorf("failed to resolve context path: %w", err)
	}

	// Resolve Dockerfile path
	dockerfilePath := buildFile
	if !filepath.IsAbs(dockerfilePath) {
		dockerfilePath = filepath.Join(absContextPath, dockerfilePath)
	}

	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("Dockerfile not found: %s", dockerfilePath)
	}

	fmt.Printf("Building from %s\n", absContextPath)
	if len(buildTags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(buildTags, ", "))
	}

	// Create tar archive of context
	buildContext, err := archive.TarWithOptions(absContextPath, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildContext.Close()

	// Build options
	buildOptions := types.ImageBuildOptions{
		Dockerfile: filepath.Base(dockerfilePath),
		Tags:       buildTags,
		NoCache:    buildNoCache,
		Remove:     true,
		Target:     buildTarget,
		BuildArgs:  make(map[string]*string),
	}

	// Parse build args
	for _, arg := range buildBuildArgs {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			buildOptions.BuildArgs[parts[0]] = &parts[1]
		}
	}

	// Build the image
	resp, err := cli.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	defer resp.Body.Close()

	// Stream build output
	fd := os.Stdout.Fd()
	isTerminal := term.IsTerminal(int(fd))
	err = jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stdout, fd, isTerminal, nil)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("\nBuild completed successfully!")

	// Push if requested
	if buildPush && len(buildTags) > 0 {
		for _, tag := range buildTags {
			fmt.Printf("Pushing %s...\n", tag)
			if err := pushImage(ctx, cli, tag); err != nil {
				return fmt.Errorf("push failed for %s: %w", tag, err)
			}
		}
	}

	return nil
}

// Push command
var dockerPushCmd = &cobra.Command{
	Use:     "push NAME[:TAG]",
	Short:   "Push an image to a registry",
	Example: `  kl docker push myapp:latest`,
	Args:    cobra.ExactArgs(1),
	RunE:    runDockerPush,
}

func runDockerPush(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	return pushImage(ctx, cli, args[0])
}

func pushImage(ctx context.Context, cli *client.Client, imageName string) error {
	// Get auth config from docker config file
	authStr := getAuthConfig(imageName)

	resp, err := cli.ImagePush(ctx, imageName, image.PushOptions{
		RegistryAuth: authStr,
	})
	if err != nil {
		return err
	}
	defer resp.Close()

	fd := os.Stdout.Fd()
	isTerminal := term.IsTerminal(int(fd))
	return jsonmessage.DisplayJSONMessagesStream(resp, os.Stdout, fd, isTerminal, nil)
}

// Pull command
var dockerPullCmd = &cobra.Command{
	Use:     "pull NAME[:TAG]",
	Short:   "Pull an image from a registry",
	Example: `  kl docker pull nginx:latest`,
	Args:    cobra.ExactArgs(1),
	RunE:    runDockerPull,
}

func runDockerPull(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	imageName := args[0]
	authStr := getAuthConfig(imageName)

	resp, err := cli.ImagePull(ctx, imageName, image.PullOptions{
		RegistryAuth: authStr,
	})
	if err != nil {
		return err
	}
	defer resp.Close()

	fd := os.Stdout.Fd()
	isTerminal := term.IsTerminal(int(fd))
	return jsonmessage.DisplayJSONMessagesStream(resp, os.Stdout, fd, isTerminal, nil)
}

// Tag command
var dockerTagCmd = &cobra.Command{
	Use:     "tag SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]",
	Short:   "Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE",
	Example: `  kl docker tag myapp:latest registry.example.com/myapp:v1.0`,
	Args:    cobra.ExactArgs(2),
	RunE:    runDockerTag,
}

func runDockerTag(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	if err := cli.ImageTag(ctx, args[0], args[1]); err != nil {
		return err
	}

	fmt.Printf("Tagged %s as %s\n", args[0], args[1])
	return nil
}

// Images command
var imagesAll bool

var dockerImagesCmd = &cobra.Command{
	Use:     "images [REPOSITORY[:TAG]]",
	Short:   "List images",
	Example: `  kl docker images`,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runDockerImages,
}

func init() {
	dockerImagesCmd.Flags().BoolVarP(&imagesAll, "all", "a", false, "Show all images")
}

func runDockerImages(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	opts := image.ListOptions{All: imagesAll}
	if len(args) > 0 {
		opts.Filters = filters.NewArgs(filters.Arg("reference", args[0]))
	}

	images, err := cli.ImageList(ctx, opts)
	if err != nil {
		return err
	}

	fmt.Printf("%-50s %-20s %-15s %s\n", "REPOSITORY", "TAG", "IMAGE ID", "SIZE")
	for _, img := range images {
		repoTags := img.RepoTags
		if len(repoTags) == 0 {
			repoTags = []string{"<none>:<none>"}
		}
		for _, tag := range repoTags {
			parts := strings.SplitN(tag, ":", 2)
			repo := parts[0]
			tagName := "latest"
			if len(parts) > 1 {
				tagName = parts[1]
			}
			imageID := img.ID
			if strings.HasPrefix(imageID, "sha256:") {
				imageID = imageID[7:19]
			}
			size := formatSize(img.Size)
			fmt.Printf("%-50s %-20s %-15s %s\n", repo, tagName, imageID, size)
		}
	}

	return nil
}

// Rmi command
var rmiForce bool

var dockerRmiCmd = &cobra.Command{
	Use:     "rmi IMAGE [IMAGE...]",
	Short:   "Remove one or more images",
	Example: `  kl docker rmi myapp:latest`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runDockerRmi,
}

func init() {
	dockerRmiCmd.Flags().BoolVarP(&rmiForce, "force", "f", false, "Force removal")
}

func runDockerRmi(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, imageName := range args {
		deleted, err := cli.ImageRemove(ctx, imageName, image.RemoveOptions{Force: rmiForce})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error removing %s: %v\n", imageName, err)
			continue
		}
		for _, d := range deleted {
			if d.Deleted != "" {
				fmt.Printf("Deleted: %s\n", d.Deleted)
			}
			if d.Untagged != "" {
				fmt.Printf("Untagged: %s\n", d.Untagged)
			}
		}
	}

	return nil
}

// Inspect command
var dockerInspectCmd = &cobra.Command{
	Use:     "inspect IMAGE [IMAGE...]",
	Short:   "Display detailed information on one or more images",
	Example: `  kl docker inspect myapp:latest`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runDockerInspect,
}

func runDockerInspect(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	var results []image.InspectResponse
	for _, imageName := range args {
		info, _, err := cli.ImageInspectWithRaw(ctx, imageName)
		if err != nil {
			return fmt.Errorf("error inspecting %s: %w", imageName, err)
		}
		results = append(results, info)
	}

	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

// Login command
var (
	loginUsername string
	loginPassword string
)

var dockerLoginCmd = &cobra.Command{
	Use:     "login [SERVER]",
	Short:   "Log in to a registry",
	Example: `  kl docker login registry.example.com`,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runDockerLogin,
}

func init() {
	dockerLoginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Username")
	dockerLoginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Password")
}

func runDockerLogin(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	serverAddress := "https://index.docker.io/v1/"
	if len(args) > 0 {
		serverAddress = args[0]
	}

	// Get username if not provided
	username := loginUsername
	if username == "" {
		fmt.Print("Username: ")
		fmt.Scanln(&username)
	}

	// Get password if not provided
	password := loginPassword
	if password == "" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = string(passwordBytes)
		fmt.Println()
	}

	authConfig := registry.AuthConfig{
		Username:      username,
		Password:      password,
		ServerAddress: serverAddress,
	}

	resp, err := cli.RegistryLogin(ctx, authConfig)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Save credentials to docker config
	if err := saveAuthConfig(serverAddress, authConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save credentials: %v\n", err)
	}

	fmt.Println(resp.Status)
	return nil
}

// Logout command
var dockerLogoutCmd = &cobra.Command{
	Use:     "logout [SERVER]",
	Short:   "Log out from a registry",
	Example: `  kl docker logout registry.example.com`,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runDockerLogout,
}

func runDockerLogout(cmd *cobra.Command, args []string) error {
	serverAddress := "https://index.docker.io/v1/"
	if len(args) > 0 {
		serverAddress = args[0]
	}

	if err := removeAuthConfig(serverAddress); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	fmt.Printf("Removed login credentials for %s\n", serverAddress)
	return nil
}

// Info command
var dockerInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display system-wide information",
	RunE:  runDockerInfo,
}

func runDockerInfo(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	info, err := cli.Info(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Server Version: %s\n", info.ServerVersion)
	fmt.Printf("Storage Driver: %s\n", info.Driver)
	fmt.Printf("Operating System: %s\n", info.OperatingSystem)
	fmt.Printf("OSType: %s\n", info.OSType)
	fmt.Printf("Architecture: %s\n", info.Architecture)
	fmt.Printf("CPUs: %d\n", info.NCPU)
	fmt.Printf("Total Memory: %s\n", formatSize(info.MemTotal))
	fmt.Printf("Docker Root Dir: %s\n", info.DockerRootDir)

	return nil
}

// Version command
var dockerVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the Docker version information",
	RunE:  runDockerVersion,
}

func runDockerVersion(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	version, err := cli.ServerVersion(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Client:\n")
	fmt.Printf(" Version:      %s\n", cli.ClientVersion())
	fmt.Printf("\nServer:\n")
	fmt.Printf(" Version:      %s\n", version.Version)
	fmt.Printf(" API version:  %s\n", version.APIVersion)
	fmt.Printf(" Go version:   %s\n", version.GoVersion)
	fmt.Printf(" Git commit:   %s\n", version.GitCommit)
	fmt.Printf(" Built:        %s\n", version.BuildTime)
	fmt.Printf(" OS/Arch:      %s/%s\n", version.Os, version.Arch)

	return nil
}

// Helper functions

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// Docker config file helpers
type dockerConfigFile struct {
	Auths map[string]dockerAuthEntry `json:"auths"`
}

type dockerAuthEntry struct {
	Auth string `json:"auth"`
}

func getDockerConfigPath() string {
	if configDir := os.Getenv("DOCKER_CONFIG"); configDir != "" {
		return filepath.Join(configDir, "config.json")
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".docker", "config.json")
}

func loadDockerConfig() (*dockerConfigFile, error) {
	configPath := getDockerConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &dockerConfigFile{Auths: make(map[string]dockerAuthEntry)}, nil
		}
		return nil, err
	}

	var config dockerConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	if config.Auths == nil {
		config.Auths = make(map[string]dockerAuthEntry)
	}
	return &config, nil
}

func saveDockerConfig(config *dockerConfigFile) error {
	configPath := getDockerConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func getAuthConfig(imageName string) string {
	config, err := loadDockerConfig()
	if err != nil {
		return ""
	}

	// Extract registry from image name
	registry := "https://index.docker.io/v1/"
	parts := strings.SplitN(imageName, "/", 2)
	if len(parts) > 1 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":")) {
		registry = parts[0]
	}

	// Try different registry URL formats
	for _, r := range []string{registry, "https://" + registry, "https://" + registry + "/v1/", "https://" + registry + "/v2/"} {
		if auth, ok := config.Auths[r]; ok {
			return auth.Auth
		}
	}

	return ""
}

func saveAuthConfig(serverAddress string, authConfig registry.AuthConfig) error {
	config, err := loadDockerConfig()
	if err != nil {
		return err
	}

	authStr := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", authConfig.Username, authConfig.Password)),
	)

	config.Auths[serverAddress] = dockerAuthEntry{Auth: authStr}
	return saveDockerConfig(config)
}

func removeAuthConfig(serverAddress string) error {
	config, err := loadDockerConfig()
	if err != nil {
		return err
	}

	delete(config.Auths, serverAddress)
	return saveDockerConfig(config)
}

// Ensure io is used
var _ = io.Copy
