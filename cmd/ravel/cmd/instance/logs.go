package instance

import (
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get logs from a instance",
	Long:  `Get logs from a instance. The instance must be running.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	InstanceCmd.AddCommand(logsCmd)
}
