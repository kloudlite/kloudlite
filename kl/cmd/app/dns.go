package app

import (
	"fmt"

	dnsserver "github.com/kloudlite/kl/cmd/app/dns-server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:    "start-dns",
	Hidden: true,
	Long:   `This is internal command`,
	Run: func(cmd *cobra.Command, _ []string) {
		addr := fn.ParseStringFlag(cmd, "addr")

		if addr == "" {
			fn.PrintError(fmt.Errorf("invalid address"))
			return
		}

		if err := dnsserver.StartDnsServer(addr); err != nil {
			fn.PrintError(err)
		}
	},
}

func init() {
	dnsCmd.Flags().StringP("addr", "a", "127.0.0.2:53", "port to listen on")
}
