package container

import (
	"context"
	"fmt"
	"time"

	"github.com/valyentdev/ravel/internal/agent/runtimes"
	"github.com/valyentdev/ravel/internal/vminit"
	"github.com/valyentdev/ravel/pkg/proto"
)

func (r *Runtime) Exec(ctx context.Context, instanceId string, cmd []string, timeout time.Duration) (*runtimes.ExecResult, error) {
	_, ok := r.runningVMs[instanceId]
	if !ok {
		return nil, fmt.Errorf("machine %q is not running", instanceId)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, client, err := vminit.NewClient(vsockPath(instanceId))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res, err := client.Exec(timeoutCtx, &proto.ExecRequest{
		Cmd: cmd,
	})
	if err != nil {
		return nil, err
	}

	return &runtimes.ExecResult{
		Stdout:   res.Output,
		ExitCode: int(res.ExitCode),
	}, nil
}
