// root.go: Defines the root command swagger-exp-knife4j, global --debug-log/--quiet flags, and Execute() error display.
package cmd

import (
	"fmt"
	"os"

	"swagger-exp-knife4j/internal/ascii"
	"swagger-exp-knife4j/pkg/config"
	"swagger-exp-knife4j/pkg/extension"
	"swagger-exp-knife4j/pkg/log"

	"github.com/spf13/cobra"
)

var (
	opts = &config.Options{}
)

var rootCmd = &cobra.Command{
	Use:   "swagger-exp-knife4j",
	Short: "A  Swagger API document testing tool",
	Long:  ascii.Logo(),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if opts.Logging.Silence {
			log.EnableSilence()
		}

		if opts.Logging.Debug && !opts.Logging.Silence {
			log.EnableDebug()
			log.Debug("debug logging enabled")
		}

		return nil
	},
}

// Execute runs the Cobra root command.
// Prints the error in a Markdown block and calls os.Exit(1) on failure.
func Execute() {
	applyAllCommandHelp()
	extension.AttachCommands(rootCmd)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SilenceErrors = true
	err := rootCmd.Execute()
	if err != nil {
		var cmd string
		c, _, cerr := rootCmd.Find(os.Args[1:])
		if cerr == nil {
			cmd = c.Name()
		}

		v := "\n"

		if cmd != "" {
			v += fmt.Sprintf("An error occured running the `%s` command\n", cmd)
		}

		v += "The error was:\n\n" + fmt.Sprintf("```%s```", err)
		fmt.Fprintln(os.Stderr, ascii.Markdown(v))

		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&opts.Logging.Debug, "debug-log", "D", false, "Show debug logging")
	rootCmd.PersistentFlags().BoolVarP(&opts.Logging.Silence, "quiet", "q", false, "Silence (almost all) logging")
}
