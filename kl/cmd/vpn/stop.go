package vpn

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"runtime"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop vpn",
	Long:  `stop vpn`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := stopVPN(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func stopVPN() error {

	if runtime.GOOS != "linux" {
		fn.Log(text.Green("stop vpn from your wireguard client"))
		return nil
	}

	//current, err := user.Current()
	//if err != nil {
	//	return fn.NewE(err)
	//}
	//
	//if current.Uid != "0" {
	//	return fn.Errorf("root permission required")
	//}
	//
	//var errBuf strings.Builder
	//cmd := exec.Command("wg-quick", "down", "kl")
	//cmd.Stderr = &errBuf
	//
	//err = cmd.Run()
	//if err != nil {
	//	return fn.Errorf(errBuf.String())
	//}

	if err := startWireguard("", true); err != nil {
		return err
	}

	fn.Log(text.Green("kloudlite vpn has been stopped"))

	return nil
}
