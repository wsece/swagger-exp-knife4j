package extension

import "github.com/spf13/cobra"

// CommandExtension adds a Cobra subcommand under the program root.
// Use pkg/log for output; read global config via flags on your command or env vars.
type CommandExtension interface {
	// CobraCommand returns a fully configured *cobra.Command (Use, Short, RunE, Flags).
	// The command is attached in cmd.Execute via AttachCommands.
	CobraCommand() *cobra.Command
}
