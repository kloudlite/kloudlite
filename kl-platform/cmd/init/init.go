package init

import (
	"slices"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kloudlite/kl-platform/domain/fileclient"
	"github.com/spf13/cobra"
)

// KeyValueMap is a custom type to handle key-value pairs without duplicate keys.
type KeyValueMap struct {
	Args []string
}

func (kv *KeyValueMap) Set(value string) error {
	// Split the input string on '='

	// Check for duplicate keys

	if slices.Contains(kv.Args, value) {
		return fn.Errorf("duplicate k3s argument: %s", value)
	}

	kv.Args = append(kv.Args, value)
	return nil
}

func (kv *KeyValueMap) String() string {
	// Convert the map back to a key=value string slice
	var pairs = kv.Args

	return strings.Join(pairs, ", ")
}

func (kv *KeyValueMap) Type() string {
	return "keyValueMap"
}

func GetCmd() *cobra.Command {
	var k3sArgs KeyValueMap
	k3sArgs.Args = make([]string, 0)

	var InitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize kloudlite platform",
		Long:  "It will create a kloudlite platform configuration file(platform-config.yml) in the current directory.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := Run(cmd, k3sArgs); err != nil {
				fn.Log(err)
			}
		},
	}

	InitCmd.Flags().Var(&k3sArgs, "k3s-args", "Set key=value pairs for k3s-args")

	InitCmd.Flags().StringP("base-domain", "b", "", "Set base domain for the platform")

	return InitCmd
}

func Run(cmd *cobra.Command, k3sArgs KeyValueMap) error {

	baseDomain := fn.ParseStringFlag(cmd, "base-domain")

	if baseDomain == "" {
		return fn.Error("base domain is required")
	}

	fc, err := fileclient.New(cmd)
	if err != nil {
		return err
	}

	args := []string{}

	cf, err := fc.GetConfigFile()
	if err == nil && len(*&cf.K3sArgs) > 0 {
		args = append(args, cf.K3sArgs...)
	}

	if len(args) == 0 {
		args = append(args, "node-name=master")
		args = append(args, "disable=traefik")
	}

	for _, v := range k3sArgs.Args {
		if slices.Contains(args, v) {
			continue
		}

		args = append(args, v)
	}

	if err := fc.WriteConfigFile(fileclient.ConfigFile{
		Version: "v1",
		K3sArgs: args,
		KlConfig: &fileclient.KlConfig{
			BaseDomain: baseDomain,
		},
	}); err != nil {
		return err
	}

	return nil
}
