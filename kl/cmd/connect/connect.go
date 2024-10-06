package connect

import (
	"bufio"
	"errors"
	"github.com/go-ping/ping"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/envclient"
	"github.com/kloudlite/kl/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"io"
	"os"
	"time"
)

var Command = &cobra.Command{
	Use:   "connect",
	Short: "start the wireguard connection",
	Long:  "This command will start the wireguard connection",
	Run: func(_ *cobra.Command, _ []string) {
		if err := startWg(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func startWg() error {
	defer spinner.Client.UpdateMessage("connecting your device")()
	k3sClient, err := k3s.NewClient()
	if err != nil {
		return fn.NewE(err)
	}

	if !envclient.InsideBox() {
		if err := k3sClient.RestartWgProxyContainer(); err != nil {
			return fn.NewE(err)
		}
		return nil
	}

	if ChekcWireguardConnection() {
		return nil
	}

	if err = fn.ExecNoOutput("wg-quick down kl-workspace-wg"); err != nil {
		return fn.NewE(err)
	}

	if err = fn.ExecNoOutput("wg-quick up kl-workspace-wg"); err != nil {
		return fn.NewE(err)
	}

	time.Sleep(time.Second * 1)

	open, err := os.Open("/tmp/kl/online.status")
	if err != nil {
		return err
	}

	if _, err := open.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	defer open.Close()
	reader := bufio.NewReader(open)

	startTime := time.Now()
	for {
		<-time.After(time.Second * 1)
		msg, err := reader.ReadString('\n')
		if err != nil {
			if time.Since(startTime) > time.Second*30 {
				return errors.New("failed to connect")
			}
			if errors.Is(err, io.EOF) {
				continue
			}
			return err
		}
		if msg == "online\n" {
			break
		}
	}
	fn.Log(text.Green("device connected"))

	return nil
}

func ChekcWireguardConnection() bool {
	pinger, err := ping.NewPinger(constants.KLDNS)
	if err != nil {
		return false
	}
	pinger.Count = 1
	pinger.Timeout = 2 * time.Second
	if err := pinger.Run(); err != nil {
		return false
	}
	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return false
	}
	return true
}
