package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strings"

	"github.com/valyentdev/ravel/cmd/ravel-init/api"
	"github.com/valyentdev/ravel/pkg/proto"

	"github.com/opencontainers/runc/libcontainer/system"
	"github.com/opencontainers/runc/libcontainer/user"
	"golang.org/x/sys/unix"
)

func main() {
	slog.Info("Starting init")
	config, err := DecodeConfig("/ravel/run.json")
	if err != nil {
		panic(err)
	}

	if err := mkdir("/dev", perm0755); err != nil {
		panic(err)
	}

	initialMnts := MakeInitialMounts(config.RootDevice)
	if err := initialMnts.Mount(); err != nil {
		panic(err)
	}

	slog.Debug("Switching root")
	if err := os.Chdir("/newroot"); err != nil {
		panic(err)
	}

	if err := mount(".", "/", "", unix.MS_MOVE, ""); err != nil {
		panic(err)
	}

	if err := unix.Chroot("."); err != nil {
		panic(err)
	}

	if err := os.Chdir("/"); err != nil {
		panic(err)
	}

	mnts := MakeMounts()
	if err := mnts.Mount(); err != nil {
		panic(err)
	}

	if err := mkdir("/run/lock", fs.FileMode(^uint32(0))); err != nil {
		panic(fmt.Errorf("could not create /run/lock directory: %w", err))
	}

	if err := unix.Symlinkat("/proc/self/fd", 0, "/dev/fd"); err != nil {
		panic(err)
	}

	if err := unix.Symlinkat("/proc/self/fd/0", 0, "/dev/stdin"); err != nil {
		panic(err)
	}

	if err := unix.Symlinkat("/proc/self/fd/1", 0, "/dev/stdout"); err != nil {
		panic(err)
	}

	if err := unix.Symlinkat("/proc/self/fd/2", 0, "/dev/stderr"); err != nil {
		panic(err)
	}

	if err := mkdir("/root", unix.S_IRWXU); err != nil {
		panic(fmt.Errorf("could not create /root dir: %w", err))
	}

	cgroupMnt := MakeCgroupMounts()
	if err := cgroupMnt.Mount(); err != nil {
		panic(err)
	}

	if err := unix.Setrlimit(0, &unix.Rlimit{Cur: 10240, Max: 10240}); err != nil {
		panic(err)
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
		panic("no username set, something is terribly wrong!")
	} else if len(usrSplit) >= 2 {
		_, err = user.LookupGroup(usrSplit[1])
		if err != nil {
			panic(fmt.Errorf("group %s not found: %v", usrSplit[1], err))
		}
	}

	nixUser, err := user.LookupUser(usrSplit[0])
	if err != nil {
		panic(fmt.Errorf("user %s not found: %w", username, err))
	}

	if err := system.Setgid(nixUser.Gid); err != nil {
		panic(fmt.Errorf("unable to set group id: %w", err))
	}

	if err := system.Setuid(nixUser.Uid); err != nil {
		panic(fmt.Errorf("unable to set group id: %w", err))
	}

	if err := PopulateProcessEnv(config.ImageConfig.Env); err != nil {
		panic(err)
	}

	if envHome := os.Getenv("HOME"); envHome == "" {
		if err := os.Setenv("HOME", nixUser.Home); err != nil {
			panic("unable to set user home directory")
		}
	}

	if err := MountAdditionalDrives(config.Mounts, nixUser.Uid, nixUser.Gid); err != nil {
		panic(fmt.Errorf("error mounting disk: %v", err))
	}

	if err := unix.Sethostname([]byte(config.Hostname)); err != nil {
		panic(fmt.Errorf("error setting hostname: %w", err))
	}

	if err := mkdir("/etc", perm0755); err != nil {
		panic(fmt.Errorf("could not create /etc dir: %w", err))
	}

	if err := os.WriteFile("/etc/hostname", []byte(config.Hostname+"\n"), perm0755); err != nil {
		panic(fmt.Errorf("error writing /etc/hostname: %w", err))
	}

	if err := WriteEtcResolv(config.EtcResolv); err != nil {
		panic(err)
	}

	if err := WriteEtcHost(config.EtcHost); err != nil {
		panic(err)
	}

	slog.Info("setting up networking")

	if err := NetworkSetup(config.Network); err != nil {
		fmt.Println(err)
		err = nil
	}

	cmd, err := buildCommand(config)
	if err != nil {
		panic(err)
	}

	slog.Info("executing command", "cmd", cmd.String())

	server := api.New(config, cmd)

	go func() {
		if err := cmd.Start(); err != nil {
			server.UpdateStatus(&proto.InitStatus{
				InitFailed: true,
			})
			return
		}

		server.UpdateStatus(&proto.InitStatus{
			ProcessStarted: true,
		})

		err := cmd.Wait()
		exitCode := int64(cmd.ProcessState.ExitCode())

		status := &proto.InitStatus{
			ProcessExited: true,
			ExitCode:      &exitCode,
		}

		if err != nil {
			errmsg := err.Error()
			status.Error = &errmsg
		}
		server.UpdateStatus(status)
	}()
	server.Serve()

}
