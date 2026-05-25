// help_style.go registers Markdown rendering for subcommand Long and Example texts via SetCommandHelp and applyAllCommandHelp.
package cmd

import (
	"swagger-exp-knife4j/internal/ascii"
	"swagger-exp-knife4j/internal/termcolor"

	"github.com/spf13/cobra"
)

// commandHelp stores the raw help text of subcommands, and wraps them with Logo/Markdown styles at runtime.
type commandHelp struct {
	long    string
	example string
}

var commandHelpRaw = map[*cobra.Command]commandHelp{}

// SetCommandHelp registers the raw long/example help text for a subcommand.
// Rendering is applied by applyAllCommandHelp before Execute.
// Parameters: long/example are Markdown or plain text; cmd is the target Cobra subcommand.
func SetCommandHelp(cmd *cobra.Command, long, example string) {
	commandHelpRaw[cmd] = commandHelp{long: long, example: example}
	applyCommandHelp(cmd)
}

func applyCommandHelp(cmd *cobra.Command) {
	raw, ok := commandHelpRaw[cmd]
	if !ok {
		return
	}
	if raw.long != "" {
		cmd.Long = ascii.LogoHelp(ascii.Markdown(raw.long))
	}
	if raw.example != "" {
		cmd.Example = ascii.Markdown(raw.example)
	}
}

func applyAllCommandHelp() {
	termcolor.ConfigureOutput()
	for cmd := range commandHelpRaw {
		applyCommandHelp(cmd)
	}
}
