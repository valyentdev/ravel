package util

import (
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/client"
)

func GetDaemonClient(cmd *cobra.Command) *client.DaemonClient {
	return client.NewDaemonClient("/var/run/ravel.sock")
}
