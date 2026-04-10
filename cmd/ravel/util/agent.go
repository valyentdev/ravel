package util

import (
	"github.com/alexisbouchez/ravel/raveld/client"
	"github.com/spf13/cobra"
)

func GetDaemonClient(cmd *cobra.Command) *client.DaemonClient {
	return client.NewDaemonClient("/var/run/ravel.sock")
}
