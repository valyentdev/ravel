package container

import (
	"context"
	"syscall"
)

func syscallSignal(signal string) syscall.Signal {
	switch signal {
	case "SIGHUP":
		return syscall.SIGHUP
	case "SIGINT":
		return syscall.SIGINT
	case "SIGQUIT":
		return syscall.SIGQUIT
	case "SIGILL":
		return syscall.SIGILL
	case "SIGTRAP":
		return syscall.SIGTRAP
	case "SIGABRT":
		return syscall.SIGABRT
	case "SIGFPE":
		return syscall.SIGFPE
	case "SIGKILL":
		return syscall.SIGKILL
	case "SIGUSR1":
		return syscall.SIGUSR1
	case "SIGSEGV":
		return syscall.SIGSEGV
	case "SIGPIPE":
		return syscall.SIGPIPE
	case "SIGALRM":
		return syscall.SIGALRM
	case "SIGTERM":
		return syscall.SIGTERM
	default:
		return 0
	}
}

func (r *Runtime) SignalVM(ctx context.Context, instanceId string, signal string) error {
	// TODO
	return nil
}
