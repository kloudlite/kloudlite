package boxpkg

import (
	"fmt"
	"time"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func (c *client) ListBox() error {

	time.Sleep(20 * time.Second)
	spinner.Client.Start("sample")

	time.Sleep(20 * time.Second)

	conts, err := c.listContainer(map[string]string{
		CONT_MARK_KEY: "true",
	})
	if err != nil && err != notFoundErr {
		return err
	}

	if err == notFoundErr {
		fn.Logf("no running container found")
		return nil
	}

	header := table.Row{table.HeaderText("container name"), table.HeaderText("path")}
	rows := make([]table.Row, 0)

	for _, a := range conts {
		rows = append(rows, table.Row{
			func() string {
				if a.Name == c.containerName {
					return text.Colored(a.Name, 2)
				}
				return a.Name
			}(),
			func() string {
				pth := fn.TrimePref(a.Labels[CONT_PATH_KEY], 50)

				if a.Name == c.containerName {
					return text.Colored(pth, 2)
				}
				return pth
			}(),
		})
	}

	fmt.Println(table.Table(&header, rows, c.cmd))

	table.TotalResults(len(conts), true)

	return nil
}
