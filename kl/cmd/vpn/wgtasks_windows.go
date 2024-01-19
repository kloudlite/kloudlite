package vpn

import (
	"errors"

	"github.com/kloudlite/kl/pkg/ui/text"
)

const (
	KlWgInterface = "wgkl"
)

func connect(verbose bool) error {
	return errors.New(
		text.Colored("This command is not available for windows, will be available soon", 209),
	)
}
func disconnect(verbose bool) error {
	return errors.New(
		text.Colored("This command is not available for windows, will be available soon", 209),
	)
}
