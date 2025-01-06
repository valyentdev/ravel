package environment

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strings"

	"os/user"

	"github.com/valyentdev/ravel/initd"
	"golang.org/x/sys/unix"
)

func (e *Env) Init() error {
	slog.Info("[ravel-initd] Starting init")
	config, err := initd.DecodeConfig("/ravel/run.json")
	if err != nil {
		return err
	}

	if err := mkdir("/dev", perm0755); err != nil {
		return err
	}

	initialMnts := makeInitialMounts(config.RootDevice)
	if err := initialMnts.Mount(); err != nil {
		return err
	}

	slog.Debug("Switching root")
	if err := os.Chdir("/newroot"); err != nil {
		return err
	}

	if err := mount(".", "/", "", unix.MS_MOVE, ""); err != nil {
		return err
	}

	if err := unix.Chroot("."); err != nil {
		return err
	}

	if err := os.Chdir("/"); err != nil {
		return err
	}

	mnts := makeMounts()
	if err := mnts.Mount(); err != nil {
		return err
	}

	if err := mkdir("/run/lock", fs.FileMode(^uint32(0))); err != nil {
		return (fmt.Errorf("could not create /run/lock directory: %w", err))
	}

	if err := unix.Symlinkat("/proc/self/fd", 0, "/dev/fd"); err != nil {
		return err
	}

	if err := unix.Symlinkat("/proc/self/fd/0", 0, "/dev/stdin"); err != nil {
		return err
	}

	if err := unix.Symlinkat("/proc/self/fd/1", 0, "/dev/stdout"); err != nil {
		return err
	}

	if err := unix.Symlinkat("/proc/self/fd/2", 0, "/dev/stderr"); err != nil {
		return err
	}

	if err := mkdir("/root", unix.S_IRWXU); err != nil {
		return (fmt.Errorf("could not create /root dir: %w", err))
	}

	cgroupMnt := makeCgroupMounts()
	if err := cgroupMnt.Mount(); err != nil {
		return err
	}

	if err := unix.Setrlimit(0, &unix.Rlimit{Cur: 10240, Max: 10240}); err != nil {
		return err
	}

	// parse user and  group names
	username := "root"
	if config.ImageConfig.User != nil {
		username = *config.ImageConfig.User
	}

	if config.UserOverride != "" {
		slog.Info("overriding user", "user", config.UserOverride)
		username = config.UserOverride
	}

	usrSplit := strings.Split(username, ":")

	if len(usrSplit) < 1 {
		return fmt.Errorf("no username set, something is terribly wrong")
	} else if len(usrSplit) >= 2 {
		_, err = user.LookupGroup(usrSplit[1])
		if err != nil {
			return (fmt.Errorf("group %s not found: %v", usrSplit[1], err))
		}
	}

	uid, gid, homeDir, err := resolveUser(usrSplit[0])
	if err != nil {
		return fmt.Errorf("error resolving user: %v", err)
	}

	if err := unix.Setgid(uid); err != nil {
		return fmt.Errorf("unable to set group id: %w", err)
	}

	if err := unix.Setuid(gid); err != nil {
		return fmt.Errorf("unable to set group id: %w", err)
	}

	if err := populateProcessEnv(config.ImageConfig.Env); err != nil {
		return err
	}

	if envHome := os.Getenv("HOME"); envHome == "" {
		if err := os.Setenv("HOME", homeDir); err != nil {
			return fmt.Errorf("unable to set user home directory")
		}
	}

	if err := mountAdditionalDrives(config.Mounts, uid, gid); err != nil {
		return (fmt.Errorf("error mounting disk: %v", err))
	}

	if err := unix.Sethostname([]byte(config.Hostname)); err != nil {
		return (fmt.Errorf("error setting hostname: %w", err))
	}

	if err := mkdir("/etc", perm0755); err != nil {
		return (fmt.Errorf("could not create /etc dir: %w", err))
	}

	if err := os.WriteFile("/etc/hostname", []byte(config.Hostname+"\n"), perm0755); err != nil {
		return (fmt.Errorf("error writing /etc/hostname: %w", err))
	}

	if err := writeEtcResolv(config.EtcResolv); err != nil {
		return (err)
	}

	if err := writeEtcHost(config.EtcHost); err != nil {
		return (err)
	}

	if err := networkSetup(config.Network); err != nil {
		slog.Error("error setting up network", "err", err)
	}

	cmd, err := buildCommand(config)
	if err != nil {
		return err
	}

	e.cmd = cmd
	e.waitCh = make(chan struct{})
	return nil
}

func (e *Env) Run() {
	result := initd.WaitResult{
		ExitCode: -1,
	}

	defer func() {
		e.result = result
		close(e.waitCh)
	}()

	slog.Info("[ravel-initd] Running command", "cmd", e.cmd.Path, "args", e.cmd.Args)
	if err := e.cmd.Start(); err != nil {
		slog.Error("[ravel-initd] error starting command", "err", err)
		return
	}

	e.cmd.Wait()
	if e.cmd.ProcessState != nil {
		result.ExitCode = e.cmd.ProcessState.ExitCode()
	}
}
