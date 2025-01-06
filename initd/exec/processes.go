package exec

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
)

func Exec(ctx context.Context, opts api.ExecOptions) (*api.ExecResult, error) {
	if len(opts.Cmd) == 0 {
		return nil, errdefs.NewInvalidArgument("cmd cannot be empty")
	}

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	timeoutCtx, cancel := context.WithTimeout(ctx, opts.GetTimeout())
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, opts.Cmd[0])

	cmd.Path = opts.Cmd[0]
	cmd.Args = opts.Cmd

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = nil

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, err
		}
	}

	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	return &api.ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
