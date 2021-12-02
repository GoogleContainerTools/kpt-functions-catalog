package docs

import (
	"bytes"

	"github.com/olekukonko/tablewriter"
)

// newMarkdownTable returns a configured md table with specified headings
func newMarkdownTable(headings []string, buf *bytes.Buffer) tablewriter.Table {
	table := tablewriter.NewWriter(buf)
	table.SetAutoFormatHeaders(false)
	table.SetHeader(headings)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAutoWrapText(false)
	return *table
}
