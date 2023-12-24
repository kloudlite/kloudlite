package table

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kloudlite/kl/lib/common/ui/color"
)

const (
	headerColor = 5
	borderColor = 238
)

func HeaderText(text string) string {
	return color.Text(text, headerColor)
}

func GetTableStyles() table.BoxStyle {
	colorReset := color.Reset()
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

func TotalResults(length int, printIt bool) string {
	return KVOutput("Total Results:", length, printIt)
}

func KVOutput(k string, v interface{}, printIt bool) string {
	result := fmt.Sprint(
		color.Text(k, headerColor), " ",
		color.Text(fmt.Sprintf("%v", v), 2),
	)

	if printIt {
		fmt.Println(result)
	}

	return result
}

func Table(header *Row, rows []Row) string {

	t := table.Table{}

	t.Style().Box = GetTableStyles()
	if header != nil {
		t.AppendHeader(getRow(*header))
	}
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
