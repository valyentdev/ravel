package instance

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/util"
)

func newStartInstanceCmd() *cobra.Command {

	var startCmd = &cobra.Command{
		Use:                   "start",
		Short:                 "Start a previously stopped instance",
		Long:                  `Start a previously stopped instance. This is a no-op if the instance is already running.`,
		DisableFlagsInUseLine: true,
		RunE:                  runStartInstance,
	}

	return startCmd
}

func runStartInstance(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.Help()
		return fmt.Errorf("please specify a instanceId")
	}

	instanceId := args[0]

	err := util.GetAgentClient(cmd).StartInstance(cmd.Context(), instanceId)
	if err != nil {
		return fmt.Errorf("unable to start instance: %w", err)
	}

	cmd.Println("Instance started")
	return nil
}
