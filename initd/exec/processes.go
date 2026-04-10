package exec

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/api/errdefs"
)

func Exec(ctx context.Context, opts api.ExecOptions) (*api.ExecResult, error) {
	if len(opts.Cmd) == 0 {
		return nil, errdefs.NewInvalidArgument("cmd cannot be empty")
	}

	name := opts.Cmd[0]
	args := opts.Cmd[1:]

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	timeoutCtx, cancel := context.WithTimeout(ctx, opts.GetTimeout())
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, name, args...)
	if cmd.Err != nil {
		return nil, errdefs.NewInvalidArgument(cmd.Err.Error())
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = nil

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, errdefs.NewInvalidArgument(err.Error())
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

// ExecOutputLine represents a single line of output from a streaming exec.
type ExecOutputLine struct {
	Stream string `json:"stream"` // "stdout" or "stderr"
	Data   string `json:"data"`
}

// ExecStreamResult is sent when the command completes.
type ExecStreamResult struct {
	ExitCode int `json:"exit_code"`
}

// ExecStream executes a command and streams output line by line.
// This is useful for long-running commands in AI sandboxes.
func ExecStream(ctx context.Context, opts api.ExecOptions, outputCh chan<- ExecOutputLine) (*ExecStreamResult, error) {
	defer close(outputCh)

	if len(opts.Cmd) == 0 {
		return nil, errdefs.NewInvalidArgument("cmd cannot be empty")
	}

	name := opts.Cmd[0]
	args := opts.Cmd[1:]

	timeoutCtx, cancel := context.WithTimeout(ctx, opts.GetTimeout())
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, name, args...)
	if cmd.Err != nil {
		return nil, errdefs.NewInvalidArgument(cmd.Err.Error())
	}

	// Get stdout and stderr pipes
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errdefs.NewUnknown("failed to create stdout pipe: " + err.Error())
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, errdefs.NewUnknown("failed to create stderr pipe: " + err.Error())
	}

	cmd.Stdin = nil

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, errdefs.NewUnknown("failed to start command: " + err.Error())
	}

	// Stream stdout and stderr concurrently
	done := make(chan struct{}, 2)

	go streamPipe(stdoutPipe, "stdout", outputCh, done)
	go streamPipe(stderrPipe, "stderr", outputCh, done)

	// Wait for both pipes to finish
	<-done
	<-done

	// Wait for command to complete
	err = cmd.Wait()
	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	// Ignore exit errors - we just care about the exit code
	if _, ok := err.(*exec.ExitError); ok {
		err = nil
	}

	return &ExecStreamResult{ExitCode: exitCode}, err
}

func streamPipe(pipe io.ReadCloser, stream string, outputCh chan<- ExecOutputLine, done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	scanner := bufio.NewScanner(pipe)
	// Increase buffer size for lines up to 1MB
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		outputCh <- ExecOutputLine{
			Stream: stream,
			Data:   scanner.Text(),
		}
	}
}
