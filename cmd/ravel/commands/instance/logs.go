package instance

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/util"
)

func newGetInstanceLogsCmd() *cobra.Command {
	var follow bool

	var getLogsCmd = &cobra.Command{
		Use:   "logs",
		Short: "Get logs of an instance",
		Long:  `Get logs of an instance`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetInstanceLogs(cmd, args, follow)
		},
	}

	getLogsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow logs")

	return getLogsCmd

}

func runGetInstanceLogs(cmd *cobra.Command, args []string, follow bool) error {
	if len(args) == 0 {
		cmd.Help()
		return fmt.Errorf("please specify a instanceId")
	}

	instanceId := args[0]

	if follow {
		return followInstanceLogs(cmd, instanceId)
	}

	return printInstanceLogs(cmd, instanceId)
}

func followInstanceLogs(cmd *cobra.Command, instanceId string) error {
	replay, logs, err := util.GetDaemonClient(cmd).SubscribeToInstanceLogs(cmd.Context(), instanceId)
	if err != nil {
		return fmt.Errorf("unable to get instance logs: %w", err)
	}

	for _, log := range replay {
		cmd.Println(log.Message)
	}

	for log := range logs {
		cmd.Println(log.Message)
	}

	cmd.Println("--- End of logs ---")

	return nil
}

func printInstanceLogs(cmd *cobra.Command, instanceId string) error {
	logs, err := util.GetDaemonClient(cmd).GetInstanceLogs(cmd.Context(), instanceId)
	if err != nil {
		return fmt.Errorf("unable to get instance logs: %w", err)
	}

	for _, log := range logs {
		cmd.Println(log.Message)
	}

	cmd.Println("--- End of logs ---")

	return nil
}
