package jailer

import (
	"flag"
	"os"
	"os/exec"
	"strconv"
)

func setupJailerFlags(config *JailerConfig) *flag.FlagSet {
	jailerFlagSet := flag.NewFlagSet("", flag.ExitOnError)
	jailerFlagSet.Bool("help h", false, "Show help")
	jailerFlagSet.IntVar(&config.Uid, "uid", 0, "UID of the process")
	jailerFlagSet.IntVar(&config.Gid, "gid", 0, "GID of the process")
	jailerFlagSet.StringVar(&config.Netns, "netns", "", "Network namespace to join")
	jailerFlagSet.StringVar(&config.NewRoot, "new-root", "", "New root directory")
	jailerFlagSet.BoolVar(&config.NewPid, "new-pid", false, "Create new PID namespace")
	jailerFlagSet.IntVar(&config.NoFiles, "rlimit-nofiles", 0, "Number of open files")
	jailerFlagSet.IntVar(&config.Fsize, "rlimit-fsize", 0, "File size limit")
	jailerFlagSet.BoolVar(&config.MountProc, "mount-proc", false, "Mount /proc inside the jail")
	jailerFlagSet.StringVar(&config.Cgroup, "cgroup", "", "CGroup to join")
	return jailerFlagSet
}

func makeCmd(jailer string, command []string, opts *options) *exec.Cmd {
	args := []string{
		jailer,
		"exec",
		"--uid", strconv.FormatInt(int64(opts.Uid), 10),
		"--gid", strconv.FormatInt(int64(opts.Gid), 10),
		"--new-root", opts.NewRoot,
	}

	if opts.netns != "" {
		args = append(args, "--netns", opts.netns)
	}

	if opts.newPidNS {
		args = append(args, "--new-pid")
	}

	if opts.mountProc {
		args = append(args, "--mount-proc")
	}

	if opts.setRlimits {
		args = append(args, "--rlimit-fsize", strconv.FormatInt(int64(opts.fsize), 10))
		args = append(args, "--rlimit-nofiles", strconv.FormatInt(int64(opts.noFiles), 10))
	}

	if opts.cgroup != "" {
		args = append(args, "--cgroup", opts.cgroup)
	}

	args = append(args, "--")
	args = append(args, command...)

	cmd := &exec.Cmd{
		Path:   jailer,
		Args:   args,
		Env:    []string{},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	return cmd
}
