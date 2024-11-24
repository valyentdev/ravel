package signals

import "syscall"

func FromString(signal string) (syscall.Signal, bool) {
	switch signal {
	case "SIGHUP":
		return syscall.SIGHUP, true
	case "SIGINT":
		return syscall.SIGINT, true
	case "SIGQUIT":
		return syscall.SIGQUIT, true
	case "SIGILL":
		return syscall.SIGILL, true
	case "SIGTRAP":
		return syscall.SIGTRAP, true
	case "SIGABRT":
		return syscall.SIGABRT, true
	case "SIGFPE":
		return syscall.SIGFPE, true
	case "SIGKILL":
		return syscall.SIGKILL, true
	case "SIGUSR1":
		return syscall.SIGUSR1, true
	case "SIGSEGV":
		return syscall.SIGSEGV, true
	case "SIGPIPE":
		return syscall.SIGPIPE, true
	case "SIGALRM":
		return syscall.SIGALRM, true
	case "SIGTERM":
		return syscall.SIGTERM, true
	default:
		return 0, false
	}
}
