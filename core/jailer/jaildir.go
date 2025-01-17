package jailer

import (
	"fmt"
	"io"
	"os"
	"path"

	"golang.org/x/sys/unix"
)

const (
	devPerms    = 0o600
	devKVMPath  = "/dev/kvm"
	devKVMMajor = 10
	devKVMMinor = 232
	devKVMMode  = unix.S_IFCHR | devPerms

	devTUNPath  = "/dev/net/tun"
	devTUNMajor = 10
	devTUNMinor = 200
	devTUNMode  = unix.S_IFCHR | devPerms

	devURandomPath  = "/dev/urandom"
	devURandomMajor = 1
	devURandomMinor = 9
	devURandomMode  = unix.S_IFCHR | devPerms
)

type jailDir struct {
	path string
	uid  int
	gid  int
}

func createJailDir(dir string, uid, gid int) (*jailDir, error) {
	err := mkdirAndChown(dir, uid, gid, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create jail directory: %w", err)
	}

	err = mkdirAndChown(path.Join(dir, "/dev"), uid, gid, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create jail /dev directory: %w", err)
	}

	return &jailDir{
		path: dir,
		uid:  uid,
		gid:  gid,
	}, nil
}

func (j *jailDir) pathInRoot(p string) string {
	return path.Join(j.path, p)
}

func (j *jailDir) MkdirAll(dir string) error {
	return mkdirAllAndChown(j.pathInRoot(dir), 0700, j.uid, j.gid)
}

func (j *jailDir) Mkdir(dir string) error {
	return mkdirAllAndChown(j.pathInRoot(dir), 0700, j.uid, j.gid)
}

func (j *jailDir) MknodAndOwn(device string, mode uint32, dev uint64) error {
	devicePath := j.pathInRoot(device)
	err := unix.Mknod(devicePath, mode, int(dev))
	if err != nil {
		return fmt.Errorf("failed to mknod %s: %w", devicePath, err)
	}

	err = unix.Chown(devicePath, j.uid, j.gid)
	if err != nil {
		return fmt.Errorf("failed to chown %s: %w", devicePath, err)
	}

	return nil
}

func (j *jailDir) AddBlockDevice(device string) error {
	stat, err := os.Stat(device)
	if err != nil {
		return err
	}
	rdev := stat.Sys().(*unix.Stat_t).Rdev

	err = j.MkdirAll(path.Dir(device))
	if err != nil {
		return err
	}

	err = unix.Mknod(device, 0600|unix.S_IFBLK, int(rdev))
	if err != nil {
		return err
	}

	err = unix.Chown(device, j.uid, j.gid)
	if err != nil {
		return err
	}

	return nil
}

func (j *jailDir) CopyFile(src, dst string, mode uint32) error {
	dstPath := j.pathInRoot(dst)
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}

	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}

	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	err = unix.Chown(dstPath, j.uid, j.gid)
	if err != nil {
		return err
	}

	err = unix.Chmod(dstPath, mode)
	if err != nil {
		return err
	}

	return nil
}

func (j *jailDir) CreateFile(path string, mode uint32) (*os.File, error) {
	pathInRoot := j.pathInRoot(path)
	file, err := os.Create(pathInRoot)
	if err != nil {
		return nil, err
	}

	err = unix.Chown(pathInRoot, j.uid, j.gid)
	if err != nil {
		return nil, err
	}

	err = unix.Chmod(pathInRoot, mode)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (j *jailDir) HardLink(src string, dst string, readonly bool) error {
	dstPath := j.pathInRoot(dst)

	err := os.Link(src, dstPath)
	if err != nil {
		return err
	}

	err = unix.Chown(dstPath, j.uid, j.gid)
	if err != nil {
		return err
	}

	if readonly {
		err = os.Chmod(dstPath, 0400)
		if err != nil {
			return err
		}
	}

	return nil
}
