package zip

import (
	"fmt"
	"os"
	"path"

	fn "github.com/kloudlite/kl/pkg/functions"
)

func Unzip(src, dest string) error {

	_ = os.RemoveAll(path.Join(dest, "kloudlite.app"))
	_ = os.RemoveAll(path.Join(dest, "__MACOSX"))

	if err := fn.ExecCmd(fmt.Sprintf("unzip %q -d %q", src, dest), nil, false); err != nil {
		return err
	}

	_ = os.RemoveAll(path.Join(dest, "__MACOSX"))

	return nil
}
