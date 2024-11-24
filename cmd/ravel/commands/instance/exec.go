package instance

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/util"
)

type execOptions struct {
	timeout time.Duration
}

func newInstanceExec() *cobra.Command {
	var execOptions execOptions

	cmd := &cobra.Command{
		Use:   "exec [instance-id] -- [command...]",
		Short: "Execute a command on a instance",
		Long:  `Execute a command on a instance.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstanceExec(cmd, args, execOptions.timeout)
		},
	}

	cmd.Flags().SetInterspersed(false)
	cmd.Flags().DurationVarP(&execOptions.timeout, "timeout", "t", 10, "Timeout")
	return cmd
}

func runInstanceExec(cmd *cobra.Command, args []string, timeout time.Duration) error {
	if len(args) < 2 {
		return fmt.Errorf("please specify a instance id, then the command")
	}

	var cmdLine []string

	if args[1] == "--" {
		println("reading from stdin")
		// Read line from stdin
		command := make([]byte, 1024)
		n, err := os.Stdin.Read(command)
		if err != nil {
			return fmt.Errorf("failed to read command from stdin: %w", err)
		}
		// Parse command

		str := string(command[:n])
		// Remove trailing newline
		if str[len(str)-1] == '\n' {
			str = str[:len(str)-1]
		}

		cmdLine = []string{
			"/bin/sh",
			"-c",
			str,
		}

	}

	instanceId := args[0]

	res, err := util.GetDaemonClient(cmd).InstanceExec(context.Background(), instanceId, cmdLine, timeout)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	_, _ = os.Stdout.Write([]byte(res.Stdout))

	return nil
}
