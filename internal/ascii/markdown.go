// markdown.go: Renders Markdown in help and error messages into terminal-readable text using glamour.
package ascii

import (
	"strings"

	"swagger-exp-knife4j/internal/termcolor"

	"github.com/charmbracelet/glamour"
)

var renderer *glamour.TermRenderer

// Markdown renders Markdown for terminal help/errors; returns raw text for non-color terminals.
func Markdown(s string) string {
	s = strings.TrimSpace(s)
	if !termcolor.Enabled {
		return s
	}
	r, err := renderer.Render(s)
	if err != nil {
		panic(err)
	}
	return r
}

func init() {
	var err error
	renderer, err = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithPreservedNewLines(),
	)
	if err != nil {
		panic(err)
	}
}
