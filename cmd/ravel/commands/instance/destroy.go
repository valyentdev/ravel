package instance

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newDestroyInstanceCmd() *cobra.Command {
	var force bool
	destroyCmd := &cobra.Command{
		Use:     "destroy",
		Aliases: []string{"rm"},
		Short:   "Remove a instance",
		Long:    `Remove a instance. The instance must be stopped.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDestroyInstance(cmd, args, force)
		},
	}

	destroyCmd.Flags().BoolVarP(&force, "force", "f", false, "Force remove instance")

	return destroyCmd
}

func runDestroyInstance(cmd *cobra.Command, args []string, force bool) error {
	if len(args) == 0 {
		cmd.Println("Please specify a instance ID")
		return fmt.Errorf("please specify a instance ID")
	}

	instanceId := args[0]

	err := GetClient(cmd).DestroyInstance(context.Background(), instanceId, force)
	if err != nil {
		return fmt.Errorf("unable to remove instance: %w", err)
	}

	cmd.Println("Instance removed")

	return nil
}
