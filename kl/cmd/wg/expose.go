package wg

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kloudlite/kl/cmd/list"
	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
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
	kl expose port -p <port>:<your_local_port> 


  # delete exposed port of selected device
	kl expose port -d -p <port>:<your_local_port> 
`,
	Run: func(_ *cobra.Command, _ []string) {
		if len(maps) == 0 {
			common.PrintError(errors.New("no port maps provided"))
			return
		}

		ports := make([]server.Port, 0)

		for _, v := range maps {
			mp := strings.Split(v, ":")
			if len(mp) != 2 {
				common.PrintError(
					errors.New("wrong map format use <server_port>:<local_port> eg: 80:3000"),
				)
				return
			}

			pp, err := strconv.ParseInt(mp[0], 10, 32)
			if err != nil {
				common.PrintError(err)
				return
			}

			tp, err := strconv.ParseInt(mp[1], 10, 32)
			if err != nil {
				common.PrintError(err)
				return
			}

			ports = append(ports, server.Port{
				Port:       int(pp),
				TargetPort: int(tp),
			})
		}

		if !deleteFlag {

			if err := server.UpdateDevice(ports, nil); err != nil {
				common.PrintError(err)
				return
			}

			fmt.Println("ports exposed")
		} else {

			if err := server.DeleteDevicePort(ports); err != nil {
				common.PrintError(err)
				return
			}

			fmt.Println("ports deleted")
		}

		list.ListDevices([]string{})

	},
}

func init() {
	exposeCmd.Flags().StringArrayVarP(
		&maps, "port", "p", []string{},
		"expose port <server_port>:<local_port>",
	)
	exposeCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "delete ports")
}
