// version.go: `swagger-exp-knife4j version` — print build metadata from internal/version.
package cmd

import (
	"fmt"

	"swagger-exp-knife4j/internal/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show build version information",
	Long:  "Print version, git commit, and build metadata defined in internal/version.",
	Run: func(cmd *cobra.Command, args []string) {
		for _, line := range version.Report() {
			fmt.Fprintln(cmd.OutOrStdout(), line)
		}
	},
}

func init() {
	SetCommandHelp(versionCmd, helpVersionLong, helpVersionExample)
	rootCmd.AddCommand(versionCmd)
}
