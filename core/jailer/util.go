package jailer

import (
	"os"

	"golang.org/x/sys/unix"
)

func mkdirAndChown(path string, uid, gid int, perm os.FileMode) error {
	err := os.Mkdir(path, perm)
	if err != nil {
		return err
	}

	err = os.Chown(path, uid, gid)
	if err != nil {
		return err
	}

	return nil
}

func mkdirAllAndChown(path string, uid, gid int, perm os.FileMode) error {
	err := os.MkdirAll(path, perm)
	if err != nil {
		return err
	}

	err = os.Chown(path, uid, gid)
	if err != nil {
		return err
	}

	return nil
}

func mkFifo(path string, uid, gid int, mode uint32) error {
	err := unix.Mkfifo(path, mode)
	if err != nil {
		return err
	}

	return os.Chown(path, uid, gid)
}
