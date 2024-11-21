package instance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/cmd/ravel/util"
)

func newStopCmd() *cobra.Command {

	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop a running instance",
		Long:  `Stop a running instance from a given config file`,
		RunE:  runStopInstance,
	}

	return stopCmd
}

func runStopInstance(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please specify a instanceId")
	}

	instanceId := args[0]

	err := util.GetAgentClient(cmd).StopInstance(context.Background(), instanceId, &api.StopConfig{})
	if err != nil {
		bytes, _ := json.Marshal(err)
		cmd.Println(string(bytes))
		return err
	}

	cmd.Println("Instance stopped")

	return nil
}
