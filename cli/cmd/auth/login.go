package auth

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/cli/pkg/config"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	clientID = "kloudlite-cli"
)

func LoginCommand() *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "Login to Kloudlite platform using device flow",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "Kloudlite gRPC server URL",
				Value:   "localhost:50061", // Default to local development
				EnvVars: []string{"KLOUDLITE_GRPC_SERVER"},
			},
			&cli.BoolFlag{
				Name:    "no-browser",
				Aliases: []string{"n"},
				Usage:   "Don't open browser automatically",
			},
		},
		Action: func(c *cli.Context) error {
			serverAddr := c.String("server")
			noBrowser := c.Bool("no-browser")

			// Connect to gRPC server
			conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return fmt.Errorf("failed to connect to server: %w", err)
			}
			defer conn.Close()

			authClient := auth.NewAuthClient(conn)
			ctx := context.Background()

			// Initiate device flow
			initResp, err := authClient.InitiateDeviceFlow(ctx, &auth.InitiateDeviceFlowRequest{
				ClientId: clientID,
			})
			if err != nil {
				return fmt.Errorf("failed to initiate device flow: %w", err)
			}

			// Display user code
			fmt.Println()
			color.Cyan("Please visit: %s", initResp.VerificationUri)
			fmt.Println()
			fmt.Printf("And enter code: ")
			color.New(color.FgYellow, color.Bold).Printf("%s\n", initResp.UserCode)
			fmt.Println()

			// Open browser if not disabled
			if !noBrowser {
				openBrowser(initResp.VerificationUriComplete)
			}

			// Start polling for authorization
			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = " Waiting for authorization..."
			s.Start()

			ticker := time.NewTicker(time.Duration(initResp.Interval) * time.Second)
			defer ticker.Stop()

			timeout := time.After(time.Duration(initResp.ExpiresIn) * time.Second)

			for {
				select {
				case <-ticker.C:
					pollResp, err := authClient.PollDeviceToken(ctx, &auth.PollDeviceTokenRequest{
						DeviceCode: initResp.DeviceCode,
						ClientId:   clientID,
					})
					if err != nil {
						s.Stop()
						return fmt.Errorf("failed to poll device token: %w", err)
					}

					if pollResp.Error != "" {
						switch pollResp.Error {
						case "authorization_pending":
							// Continue polling
							continue
						case "slow_down":
							// Increase polling interval
							ticker.Reset(time.Duration(initResp.Interval*2) * time.Second)
							continue
						case "expired_token":
							s.Stop()
							return fmt.Errorf("device code expired, please try again")
						case "access_denied":
							s.Stop()
							return fmt.Errorf("access denied")
						default:
							s.Stop()
							return fmt.Errorf("authentication error: %s", pollResp.Error)
						}
					}

					if pollResp.Authorized {
						s.Stop()
						
						// Save tokens to config
						cfg := &config.Config{
							AccessToken:  pollResp.Token,
							RefreshToken: pollResp.RefreshToken,
							UserID:       pollResp.UserId,
							ServerAddr:   serverAddr,
						}
						
						if err := config.Save(cfg); err != nil {
							return fmt.Errorf("failed to save config: %w", err)
						}

						color.Green("✓ Successfully logged in!")
						return nil
					}

				case <-timeout:
					s.Stop()
					return fmt.Errorf("authentication timed out")
				}
			}
		},
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Start()
}