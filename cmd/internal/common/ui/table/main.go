package table

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"kloudlite.io/cmd/internal/common/ui/color"
)

const (
	headerColor = 4
	borderColor = 238
)

func HeaderText(text string) string {
	return color.ColorText(text, headerColor)
}

func GetTableStyles() table.BoxStyle {
	colorReset := color.ColorReset()
	colorBorder := color.Color(borderColor)

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

func TotalResults(length int) string {
	return KVOutput("Total Results:", length)
}

func KVOutput(k string, v interface{}) string {
	return fmt.Sprint(
		color.ColorText(k, 4), " ",
		color.ColorText(fmt.Sprintf("%v", v), 2),
	)
}

func Table(header Row, rows []Row) string {

	t := table.Table{}

	t.Style().Box = GetTableStyles()
	t.AppendHeader(getRow(header))
	t.AppendRows(getRows(rows))

	return t.Render()

	// fmt.Println(t.Render())
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
