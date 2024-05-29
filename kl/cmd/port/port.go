package port

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/fwd"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Hidden:  true,
	Use:     "port",
	Short:   "utility to check, update ports exposed from the box",
	Example: fn.Desc("{cmd} status"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := PortActions(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func PortActions(cmd *cobra.Command, _ []string) error {

	maps, err := cmd.Flags().GetStringArray("port")
	if err != nil {
		return err
	}

	p, err := proxy.NewProxy(true)
	if err != nil {
		return err
	}

	list := fn.ParseBoolFlag(cmd, "list")

	if list {
		p.ListPorts(fwd.StartCh{})
		return nil
	}

	if len(maps) != 0 && !client.InsideBox() {
		return fmt.Errorf("updating exposed ports is only allowed inside the box")
	}

	isDelete := fn.ParseBoolFlag(cmd, "delete")

	ports := make([]server.AppPort, 0)

	for _, v := range maps {
		mp := strings.Split(v, ":")
		if len(mp) != 2 {
			return errors.New("wrong map format use <server_port>:<local_port> eg: 80:3000")
		}

		pp, err := strconv.ParseInt(mp[0], 10, 32)
		if err != nil {
			return err
		}

		tp, err := strconv.ParseInt(mp[1], 10, 32)
		if err != nil {
			return err
		}

		ports = append(ports, server.AppPort{
			AppPort:    int(pp),
			DevicePort: int(tp),
		})
	}

	sshPort, ok := os.LookupEnv("SSH_PORT")
	if !ok {
		return errors.New("SSH_PORT not set")
	}

	var chMsg []fwd.StartCh

	for _, ap := range ports {
		chMsg = append(chMsg, fwd.StartCh{
			RemotePort: fmt.Sprint(ap.DevicePort),
			SshPort:    sshPort,
			LocalPort:  fmt.Sprint(ap.DevicePort),
		})
	}

	if len(ports) == 0 {
		fn.PrintError(fmt.Errorf("no ports specified"))
		cmd.Help()

		return nil
	}

	if isDelete {
		_, err := p.RemoveFwd(chMsg)
		if err != nil {
			return err
		}

		fn.Log("removed ports from forwarding")
	} else {
		_, err := p.AddFwd(chMsg)
		if err != nil {
			return err
		}

		fn.Log("added ports ot forwarding")
	}

	return nil
}

func init() {
	Cmd.Flags().BoolP("list", "l", false, "list all exposed ports")
	Cmd.Flags().BoolP("delete", "d", false, "delete port mapping")
	Cmd.Flags().StringArrayP(
		"port", "p", []string{},
		"expose port <server_port>:<local_port> outside the box",
	)
}
