package table

import (
	"encoding/json"
	"fmt"

	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	headerColor = 5
	borderColor = 238
)

func HeaderText(txt string) string {
	return text.Colored(txt, headerColor)
}

func GetTableStyles() table.BoxStyle {
	colorReset := text.Reset()
	colorBorder := text.Color(borderColor)

	return table.BoxStyle{
		BottomLeft:       colorBorder + "└",
		BottomRight:      "┘" + colorReset,
		BottomSeparator:  "┴",
		EmptySeparator:   "  ",
		Left:             colorBorder + "│" + colorReset,
		LeftSeparator:    colorBorder + "├",
		MiddleHorizontal: "─",
		MiddleSeparator:  "┼",
		MiddleVertical:   colorBorder + "│" + colorReset,
		PaddingLeft:      "  ",
		PaddingRight:     "  ",
		PageSeparator:    "  ",
		Right:            colorBorder + "│" + colorReset,
		RightSeparator:   "┤" + colorReset,
		TopLeft:          colorBorder + "┌",
		TopRight:         "┐" + colorReset,
		TopSeparator:     "┬",
		UnfinishedRow:    "  ",
	}

}

func TotalResults(length int, printIt bool) string {
	return KVOutput("Total Results:", length, printIt)
}

func KVOutput(k string, v interface{}, printIt bool) string {
	result := fmt.Sprint(
		text.Bold(k), " ",
		text.Colored(fmt.Sprintf("%v", v), 2),
	)

	if printIt {
		fmt.Println(result)
	}

	return result
}

func Table(header *Row, rows []Row, cmds ...*cobra.Command) string {
	t := table.Table{}

	output := "table"
	if len(cmds) > 0 && header != nil {
		cmd := cmds[0]
		if cmd.Flags().Changed("output") {
			output, _ = cmd.Flags().GetString("output")
		}
	}

	switch output {
	case "json":
		jsonVal := []map[string]interface{}{}

		// row is array of interface
		for _, row := range rows {
			// r is interface
			r := map[string]interface{}{}
			for i, v := range row {
				r[text.Plain((*header)[i].(string))] = text.Plain(v.(string))
			}
			jsonVal = append(jsonVal, r)
		}

		b, err := json.Marshal(jsonVal)
		if err != nil {
			return fmt.Sprint(err)
		}

		return string(b)

	case "yaml", "yml":
		jsonVal := []map[string]interface{}{}

		// row is array of interface
		for _, row := range rows {
			// r is interface
			r := map[string]interface{}{}
			for i, v := range row {
				r[text.Plain((*header)[i].(string))] = text.Plain(v.(string))
			}
			jsonVal = append(jsonVal, r)
		}

		b, err := yaml.Marshal(jsonVal)
		if err != nil {
			return fmt.Sprint(err)
		}

		return string(b)

	default:
		t.Style().Box = GetTableStyles()
		if header != nil {
			t.AppendHeader(getRow(*header))
		}
		t.AppendRows(getRows(rows))

		return t.Render()
	}
}

type Row []interface {
}

func getRow(r Row) table.Row {
	return table.Row(r)
}

func getRows(r []Row) []table.Row {
	t := make([]table.Row, 0)
	for _, r2 := range r {
		t = append(t, getRow(r2))
	}
	return t
}
