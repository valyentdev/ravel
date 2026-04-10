package jailer

import (
	"fmt"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

func chroot(path string) error {
	rootDir := "/"

	err := unix.Mount("", rootDir, "", syscall.MS_SLAVE|syscall.MS_REC, "")
	if err != nil {
		return fmt.Errorf("failed to mount: %w", err)
	}

	err = unix.Mount(path, path, "", syscall.MS_BIND|syscall.MS_REC, "")
	if err != nil {
		return err
	}

	err = syscall.Chdir(path)
	if err != nil {
		return fmt.Errorf("failed to chdir: %w", err)
	}

	err = syscall.Mkdir("old_root", 0755)
	if err != nil {
		return fmt.Errorf("failed to create old_root: %w", err)
	}

	oldRootAbs, err := filepath.Abs("./old_root")
	if err != nil {
		return fmt.Errorf("failed to get absolute path of old root: %w", err)
	}

	err = syscall.PivotRoot(path, oldRootAbs)
	if err != nil {
		return fmt.Errorf("failed to pivot_root: %w", err)
	}

	err = syscall.Chdir("/")
	if err != nil {
		return fmt.Errorf("failed to chdir: %w", err)
	}

	err = syscall.Unmount("old_root", syscall.MNT_DETACH)
	if err != nil {
		return fmt.Errorf("failed to unmount old_root: %w", err)
	}

	err = syscall.Rmdir("old_root")
	if err != nil {
		return fmt.Errorf("failed to remove old_root: %w", err)
	}

	return nil
}
