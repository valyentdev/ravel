package instance

import (
	"github.com/spf13/cobra"
)

// instanceCmd represents the instance command
var InstanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Manage ravel instances",
	Long:  `Commands to manage ravel instances.`,
}
