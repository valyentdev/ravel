package environment

import (
	"os/exec"

	"github.com/valyentdev/ravel/initd"
	"golang.org/x/sys/unix"
)

type Env struct {
	cmd    *exec.Cmd
	waitCh chan struct{}
	result initd.WaitResult
	uid    int
	gid    int
}

func (e *Env) Wait() initd.WaitResult {
	<-e.waitCh
	return e.result
}

func (e *Env) Signal(sig int) error {
	return e.cmd.Process.Signal(unix.Signal(sig))
}
