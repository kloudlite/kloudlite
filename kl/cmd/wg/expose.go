package wg

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var maps []string
var deleteFlag bool

var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "expose port of selected device",
	Long: `expose port
Examples:
  # expose port of selected device
	kl wg expose port -p <port>:<your_local_port> 


  # delete exposed port of selected device
	kl wg expose port -d -p <port>:<your_local_port> 
`,
	Run: func(_ *cobra.Command, _ []string) {
		if len(maps) == 0 {
			fn.PrintError(errors.New("no port maps provided"))
			return
		}

		ports := make([]server.DevicePort, 0)

		for _, v := range maps {
			mp := strings.Split(v, ":")
			if len(mp) != 2 {
				fn.PrintError(
					errors.New("wrong map format use <server_port>:<local_port> eg: 80:3000"),
				)
				return
			}

			pp, err := strconv.ParseInt(mp[0], 10, 32)
			if err != nil {
				fn.PrintError(err)
				return
			}

			tp, err := strconv.ParseInt(mp[1], 10, 32)
			if err != nil {
				fn.PrintError(err)
				return
			}

			ports = append(ports, server.DevicePort{
				Port:       int(pp),
				TargetPort: int(tp),
			})
		}

		if !deleteFlag {
			if err := server.UpdateDevice(ports); err != nil {
				fn.PrintError(err)
				return
			}

			fmt.Println("ports exposed")
		} else {
			if err := server.DeleteDevicePort(ports); err != nil {
				fn.PrintError(err)
				return
			}

			fmt.Println("ports deleted")
		}

	},
}

func init() {
	exposeCmd.Flags().StringArrayVarP(
		&maps, "port", "p", []string{},
		"expose port <server_port>:<local_port>",
	)
	exposeCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "delete ports")
}
