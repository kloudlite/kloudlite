package auth

import (
	"fmt"
	"github.com/kloudlite/kl/cmd/use"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to kloudlite",
	Run: func(cmd *cobra.Command, _ []string) {
		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		_, err = fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		loginId, err := apic.CreateRemoteLogin()
		if err != nil {
			fn.PrintError(err)
			return
		}

		link := fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId)

		fn.Log(text.Colored("Opening browser for login in the browser to authenticate your account\n", 2))
		fn.Println(text.Colored(text.Blue(link), 21))
		fn.Log("\n")

		//go func() {
		//	fn.Log("press enter to open link in browser")
		//	reader, err := bufio.NewReader(os.Stdin).ReadString('\n')
		//	if err != nil {
		//		fn.PrintError(err)
		//		return
		//	}
		//	if strings.Contains(reader, "\n") {
		//		err := fn.OpenUrl(link)
		//		if err != nil {
		//			fn.PrintError(err)
		//			return
		//		}
		//	} else {
		//		fn.Log("Invalid input\n")
		//	}
		//}()

		if err = apic.Login(loginId); err != nil {
			fn.PrintError(err)
			return
		}

		extraData, err := fileclient.GetExtraData()
		if err != nil {
			fn.PrintError(err)
			return
		}

		HostDNSSuffix, err := apic.GetHostDNSSuffix()
		if err != nil {
			fn.PrintError(err)
			return
		}
		extraData.DnsHostSuffix = HostDNSSuffix
		err = fileclient.SaveExtraData(extraData)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err = use.UseTeam(cmd); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("successfully logged in\n")
	},
}
