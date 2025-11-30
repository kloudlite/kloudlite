package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	buildTags       []string
	buildFile       string
	buildPush       bool
	buildNoCache    bool
	buildBuildArgs  []string
	buildTarget     string
	buildOutputType string
	buildOutputDest string
)

var buildCmd = &cobra.Command{
	Use:   "build [OPTIONS] PATH",
	Short: "Build container images using BuildKit",
	Long: `Build container images from a Dockerfile using the BuildKit daemon.

The build command connects to the BuildKit daemon running in your workspace
namespace and builds container images. It supports all standard Dockerfile
instructions and can push images directly to registries.

The BUILDKIT_HOST environment variable is automatically set in workspace
containers to point to the BuildKit daemon.`,
	Example: `  # Build an image with a tag
  kl build -t myapp:latest .

  # Build and push to a registry
  kl build -t registry.example.com/myapp:latest --push .

  # Build with a custom Dockerfile
  kl build -t myapp:latest -f Dockerfile.prod .

  # Build with build arguments
  kl build -t myapp:latest --build-arg VERSION=1.0.0 .

  # Build a specific target stage
  kl build -t myapp:latest --target builder .`,
	Args: cobra.ExactArgs(1),
	RunE: runBuild,
}

func init() {
	buildCmd.Flags().StringArrayVarP(&buildTags, "tag", "t", nil, "Name and optionally a tag (format: name:tag)")
	buildCmd.Flags().StringVarP(&buildFile, "file", "f", "Dockerfile", "Name of the Dockerfile")
	buildCmd.Flags().BoolVar(&buildPush, "push", false, "Push the image to the registry after building")
	buildCmd.Flags().BoolVar(&buildNoCache, "no-cache", false, "Do not use cache when building the image")
	buildCmd.Flags().StringArrayVar(&buildBuildArgs, "build-arg", nil, "Set build-time variables")
	buildCmd.Flags().StringVar(&buildTarget, "target", "", "Set the target build stage to build")
	buildCmd.Flags().StringVar(&buildOutputType, "output-type", "", "Output type (image, local, tar)")
	buildCmd.Flags().StringVar(&buildOutputDest, "output-dest", "", "Output destination path (for local/tar output types)")

	RootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get BuildKit host from environment
	buildkitHost := os.Getenv("BUILDKIT_HOST")
	if buildkitHost == "" {
		return fmt.Errorf("BUILDKIT_HOST environment variable not set. Are you running inside a Kloudlite workspace?")
	}

	// Get context path
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

	// Check if Dockerfile exists
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("Dockerfile not found: %s", dockerfilePath)
	}

	// Validate tags
	if len(buildTags) == 0 && buildPush {
		return fmt.Errorf("--tag is required when using --push")
	}

	fmt.Printf("Building from %s\n", absContextPath)
	if len(buildTags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(buildTags, ", "))
	}

	// Connect to BuildKit daemon
	c, err := client.New(ctx, buildkitHost)
	if err != nil {
		return fmt.Errorf("failed to connect to BuildKit daemon at %s: %w", buildkitHost, err)
	}
	defer c.Close()

	// Prepare solve options
	solveOpt := client.SolveOpt{
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			"filename": filepath.Base(dockerfilePath),
		},
		LocalDirs: map[string]string{
			"context":    absContextPath,
			"dockerfile": filepath.Dir(dockerfilePath),
		},
		Session: []session.Attachable{
			authprovider.NewDockerAuthProvider(authprovider.DockerAuthProviderConfig{}),
		},
	}

	// Set build arguments
	for _, arg := range buildBuildArgs {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			solveOpt.FrontendAttrs["build-arg:"+parts[0]] = parts[1]
		}
	}

	// Set target stage if specified
	if buildTarget != "" {
		solveOpt.FrontendAttrs["target"] = buildTarget
	}

	// Set no-cache if specified
	if buildNoCache {
		solveOpt.FrontendAttrs["no-cache"] = ""
	}

	// Configure output based on flags
	if buildOutputType != "" {
		// Custom output type specified
		switch buildOutputType {
		case "local":
			if buildOutputDest == "" {
				return fmt.Errorf("--output-dest is required when using --output-type=local")
			}
			solveOpt.Exports = []client.ExportEntry{
				{
					Type:      client.ExporterLocal,
					OutputDir: buildOutputDest,
				},
			}
		case "tar":
			if buildOutputDest == "" {
				return fmt.Errorf("--output-dest is required when using --output-type=tar")
			}
			f, err := os.Create(buildOutputDest)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer f.Close()
			solveOpt.Exports = []client.ExportEntry{
				{
					Type:   client.ExporterTar,
					Output: fixedWriteCloser(f),
				},
			}
		case "image":
			// Will be handled below with tags
		default:
			return fmt.Errorf("unsupported output type: %s", buildOutputType)
		}
	}

	// Configure image output with tags
	if len(buildTags) > 0 && (buildOutputType == "" || buildOutputType == "image") {
		exportAttrs := map[string]string{
			"name": strings.Join(buildTags, ","),
		}
		if buildPush {
			exportAttrs["push"] = "true"
		}
		solveOpt.Exports = []client.ExportEntry{
			{
				Type:  client.ExporterImage,
				Attrs: exportAttrs,
			},
		}
	}

	// Create progress display
	eg, ctx := errgroup.WithContext(ctx)
	ch := make(chan *client.SolveStatus)

	eg.Go(func() error {
		_, err := c.Solve(ctx, nil, solveOpt, ch)
		return err
	})

	eg.Go(func() error {
		// Display progress
		display, err := progressui.NewDisplay(os.Stderr, progressui.AutoMode)
		if err != nil {
			return err
		}
		_, err = display.UpdateFrom(ctx, ch)
		return err
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("\nBuild completed successfully!")
	if buildPush && len(buildTags) > 0 {
		fmt.Printf("Pushed: %s\n", strings.Join(buildTags, ", "))
	}

	return nil
}

// fixedWriteCloser wraps a file to implement the WriteCloser interface
func fixedWriteCloser(f *os.File) func(map[string]string) (io.WriteCloser, error) {
	return func(map[string]string) (io.WriteCloser, error) {
		return f, nil
	}
}

// Placeholder for LLB-based builds (for future use)
func buildWithLLB(ctx context.Context, c *client.Client) error {
	// Example of building with LLB directly
	state := llb.Image("alpine:latest").
		Run(llb.Shlex("apk add --no-cache curl")).
		Root()

	def, err := state.Marshal(ctx, llb.LinuxAmd64)
	if err != nil {
		return err
	}

	ch := make(chan *client.SolveStatus)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		_, err := c.Solve(ctx, def, client.SolveOpt{}, ch)
		return err
	})

	eg.Go(func() error {
		display, err := progressui.NewDisplay(os.Stderr, progressui.AutoMode)
		if err != nil {
			return err
		}
		_, err = display.UpdateFrom(ctx, ch)
		return err
	})

	return eg.Wait()
}
