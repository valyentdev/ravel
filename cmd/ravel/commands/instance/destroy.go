package instance

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/util"
)

func newDestroyInstanceCmd() *cobra.Command {
	destroyCmd := &cobra.Command{
		Use:     "destroy <instance_id>",
		Aliases: []string{"rm"},
		Short:   "Remove a instance",
		Long:    `Remove a instance. The instance must be stopped.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDestroyInstance(cmd, args)
		},
	}

	return destroyCmd
}

func runDestroyInstance(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.Println("Please specify a instance ID")
		return fmt.Errorf("please specify a instance ID")
	}

	instanceId := args[0]

	err := util.GetDaemonClient(cmd).DestroyInstance(context.Background(), instanceId)
	if err != nil {
		return fmt.Errorf("unable to remove instance: %w", err)
	}

	cmd.Println("Instance destroyed")

	return nil
}
