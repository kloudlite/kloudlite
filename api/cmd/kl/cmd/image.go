package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Manage container images",
	Long:  `Commands for building and managing container images with the Kloudlite registry.`,
}

var imagePushCmd = &cobra.Command{
	Use:   "push [docker-build-args...] <image-name>",
	Short: "Build and push an image to the Kloudlite registry",
	Long: `Build a Docker image and push it to the Kloudlite image registry.

This command wraps 'docker build' and automatically tags the image with the
Kloudlite registry URL and your username.

The image will be tagged as: $KL_IMAGE_REGISTRY/{username}/{image-name}

All docker build arguments are supported and passed through to docker build.
The build context defaults to the current directory if not specified.`,
	Example: `  # Build and push an image from current directory
  kl image push myapp

  # Build with a specific Dockerfile
  kl image push -f Dockerfile.prod myapp

  # Build with build args
  kl image push --build-arg VERSION=1.0 myapp

  # Build from a specific context
  kl image push -f ./docker/Dockerfile . myapp

  # Build with a tag that includes version
  kl image push myapp:v1.0.0`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true, // Pass all args to docker build
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleImagePush(args)
	},
}

func init() {
	imageCmd.AddCommand(imagePushCmd)
	RootCmd.AddCommand(imageCmd)
}

func handleImagePush(args []string) error {
	// Get the image registry from environment
	registry := os.Getenv("KL_IMAGE_REGISTRY")
	if registry == "" {
		return fmt.Errorf("KL_IMAGE_REGISTRY environment variable is not set\nThis command should be run inside a Kloudlite workspace")
	}

	// Get the username from WORKSPACE_OWNER (set in /etc/environment)
	username := os.Getenv("WORKSPACE_OWNER")
	if username == "" {
		return fmt.Errorf("WORKSPACE_OWNER environment variable is not set\nThis command should be run inside a Kloudlite workspace")
	}

	// The last argument is the image name
	imageName := args[len(args)-1]
	dockerArgs := args[:len(args)-1]

	// If the image name starts with '-', it's probably a flag
	if strings.HasPrefix(imageName, "-") {
		return fmt.Errorf("image name is required as the last argument\nUsage: kl image push [docker-build-args...] <image-name>")
	}

	// Construct the full image tag
	fullImageTag := fmt.Sprintf("%s/%s/%s", registry, username, imageName)

	// Build docker buildx command args (modern BuildKit-based builder)
	// Using buildx with --push to build and push in one step
	buildArgs := []string{"buildx", "build", "--push"}
	buildArgs = append(buildArgs, dockerArgs...)
	buildArgs = append(buildArgs, "-t", fullImageTag)

	// Check if context is provided in args (a path that doesn't start with -)
	// Look for positional args that look like paths
	hasContext := false
	for i, arg := range dockerArgs {
		if !strings.HasPrefix(arg, "-") {
			// Check if previous arg was a flag that takes a value
			if i > 0 {
				prevArg := dockerArgs[i-1]
				// Common flags that take values
				if prevArg == "-f" || prevArg == "--file" ||
					prevArg == "-t" || prevArg == "--tag" ||
					prevArg == "--build-arg" || prevArg == "--target" ||
					prevArg == "--platform" || prevArg == "--cache-from" ||
					prevArg == "--label" || prevArg == "--secret" ||
					prevArg == "--ssh" || prevArg == "-o" || prevArg == "--output" {
					continue
				}
			}
			// This looks like a context path
			hasContext = true
			break
		}
	}

	// Add default context if not provided
	if !hasContext {
		buildArgs = append(buildArgs, ".")
	}

	fmt.Printf("[+] Building and pushing image: %s\n", fullImageTag)
	fmt.Printf("[+] Running: docker %s\n\n", strings.Join(buildArgs, " "))

	// Execute docker buildx build --push
	buildCmd := exec.Command("docker", buildArgs...)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	buildCmd.Stdin = os.Stdin

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("docker buildx build failed: %w", err)
	}

	fmt.Printf("\n[✓] Image built and pushed successfully: %s\n", fullImageTag)
	return nil
}
