package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log/slog"
	// "net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"sigs.k8s.io/yaml"
)

var (
	logger  = slog.New(slog.NewTextHandler(os.Stdout, nil))
	BuiltAt string
)

func execK3s(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "k3s", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// logger.Info("STARTING k3s @ timestamp: %s, with cmd: %s\n", time.Now().Format(time.RFC3339), cmd.String())
	logger.Info("STARTING k3s", "timestamp", time.Now().Format(time.RFC3339), "cmd", cmd.String())

	if err := cmd.Run(); err != nil {
		logger.Error("while executing command, got", "err", err)
		return err
	}

	return cmd.Wait()
}

// func getPublicIPv4() (string, error) {
// 	req, err := http.NewRequest(http.MethodGet, "https://ifconfig.me", nil)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	r, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	b, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer r.Body.Close()
//
// 	return string(b), nil
// }

func main() {
	var runnerCfgFile string
	var hasVersionFlag bool
	flag.StringVar(&runnerCfgFile, "config", "./runner-config.yml", "--config runner-config-file")
	flag.BoolVar(&hasVersionFlag, "version", false, "--version")
	flag.Parse()

	if hasVersionFlag {
		logger.Info("kloudlite k3s runner", "built-at", BuiltAt)
		os.Exit(0)
	}

	logger.Info("kloudlite k3s runner", "built-at", BuiltAt)
	logger.Info("using", "configuration-file", runnerCfgFile)

	ctx, cf := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	defer cf()

	for {
		if err := ctx.Err(); err != nil {
			logger.Error("command context cancelled")
			os.Exit(1)
		}
		f, err := os.Open(runnerCfgFile)
		if err != nil {
			logger.Error("failed to open configuration file, encountered", "err", err)
			<-time.After(3 * time.Second)
			continue
		}

		logger.Info("found runner config file")

		out, err := io.ReadAll(f)
		if err != nil {
			logger.Error("failed to open configuration file, encountered", "err", err)
			continue
		}

		var runnerCfg K3sRunnerConfig
		if err := yaml.Unmarshal(out, &runnerCfg); err != nil {
			logger.Error("while unmarshaling, got", "err", err)
			continue
		}

		if err := execK3s(ctx, runnerCfg.K3sFlags...); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error("failed to start k3s, encountered", "err", err)
				logger.Info("will retry in 10 seconds")
				<-time.After(10 * time.Second)
				continue
			}
		}

		logger.Info("successfully started runner", "timestamp", time.Now().Format(time.RFC3339))
		break
	}
}
