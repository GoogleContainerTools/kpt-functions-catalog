package docs

import (
	"fmt"
	"strings"
)

func getMdLink(d, link string) string {
	return fmt.Sprintf("[%s](%s)", d, link)
}

func getMdListItem(d string) string {
	return fmt.Sprintf("- %s", d)
}

func getMdHeading(title string, l int) string {
	return fmt.Sprintf("%s %s\n\n", strings.Repeat("#", l), title)
}
