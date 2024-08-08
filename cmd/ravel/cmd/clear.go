package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove all ravel data from the system",
	Long:  `Remove all ravel data from the system`,
	Run: func(cmd *cobra.Command, args []string) {
		err := os.RemoveAll("/var/lib/ravel")
		if err != nil {
			panic(err)
		}

		err = os.RemoveAll("/var/log/ravel")
		if err != nil {
			panic(err)
		}

		err = os.RemoveAll("/tmp/ravel")
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
