package instance

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	raveldclient "github.com/valyentdev/ravel/cmd/ravel/client"
	"github.com/valyentdev/ravel/pkg/proto"
)

func init() {
	cmd := &cobra.Command{
		Use:   "exec [instance-id] -- [command...]",
		Short: "Execute a command on a instance",
		Long:  `Execute a command on a instance.`,
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	timeout := flags.Int64P("timeout", "t", 10000, "Timeout")
	env := flags.StringSliceP("env", "e", []string{}, "Env")
	workingDir := flags.StringP("directory", "w", "", "Working directory")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.Println("Please specify a instance id, then the command")
			cmd.Help()
			return
		}

		var cmdLine []string

		if args[1] == "--" {
			println("reading from stdin")
			// Read line from stdin
			command := make([]byte, 1024)
			n, err := os.Stdin.Read(command)
			if err != nil {
				cmd.Println("Failed to read command from stdin: ", err)
				return
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

		req := &proto.InstanceExecRequest{
			InstanceId: args[0],
			ExecRequest: &proto.ExecRequest{
				Cmd:        args[1:],
				Env:        *env,
				WorkingDir: nil,
			},
			Timeout: *timeout,
		}

		if cmdLine != nil {
			req.ExecRequest.Cmd = cmdLine
		}

		if *workingDir != "" {
			req.ExecRequest.WorkingDir = workingDir
		}
		res, err := raveldclient.DaemonClient.InstanceExec(context.Background(), req)
		if err != nil {
			cmd.Println("Failed to execute command: ", err)
			return
		}

		_, _ = os.Stdout.Write(res.Output)
		os.Exit(int(res.ExitCode))
	}
	InstanceCmd.AddCommand(cmd)
}
