package use

import (
	"errors"
	"fmt"

	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var regionCmd = &cobra.Command{
	Use:   "region",
	Short: "select region of selected device",
	Long: `Select region
Examples:
  # select region of selected device
  kl use region
	`,
	Run: func(_ *cobra.Command, args []string) {
		regionArg := ""
		if len(args) >= 1 {
			regionArg = args[0]
		}

		region, err := SelectRegion(regionArg)

		if err != nil {
			common_util.PrintError(err)
			return
		}

		fmt.Println(region)

		if err = server.UpdateDevice([]server.Port{}, &region); err != nil {
			common_util.PrintError(err)
			return
		}

		fmt.Println("successfully selected region")

	},
}

func SelectRegion(region string) (string, error) {

	regions, err := server.GetRegions()
	if err != nil {
		return "", err
	}

	if len(regions) == 0 {
		return "", errors.New("no regions available in current account")
	}

	if region != "" {
		for _, r := range regions {
			if r.Region == region {
				return region, nil
			}
		}

		return "", errors.New("provider region not available in your selected account")
	}

	selectedIndex, err := fuzzyfinder.Find(
		regions,
		func(i int) string {
			return fmt.Sprintf("%s, %s", regions[i].Name, regions[i].Region)
		},
		fuzzyfinder.WithPromptString("Select Device >"),
	)

	if err != nil {
		return "", err
	}

	return regions[selectedIndex].Region, nil
}

func init() {
}
