package example

import (
	"github.com/spf13/cobra"
	"swagger-exp-knife4j/pkg/log"
)

type helloCommand struct{}

func (helloCommand) CobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "hello-example",
		Short: "Example extension command (from extensions/example)",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			log.Print("hello from extensions/example")
			return nil
		},
	}
}
