package jailer

import (
	"os"
	"path/filepath"
	"syscall"

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

func mkFifo(path string, uid, gid int, mode uint32) error {
	err := unix.Mkfifo(path, mode)
	if err != nil {
		return err
	}

	return os.Chown(path, uid, gid)
}

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the https://cs.opensource.google/go/go LICENSE file.
//
// Extracted from os.MkdirAll
func mkdirAllAndChown(path string, perm os.FileMode, uid, gid int) error {
	// Fast path: if we can tell whether path is a directory or file, stop with success or error.
	dir, err := os.Stat(path)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return &os.PathError{Op: "mkdir", Path: path, Err: syscall.ENOTDIR}
	}

	// Slow path: make sure parent exists and then call Mkdir for path.

	// Extract the parent folder from path by first removing any trailing
	// path separator and then scanning backward until finding a path
	// separator or reaching the beginning of the string.
	i := len(path) - 1
	for i >= 0 && os.IsPathSeparator(path[i]) {
		i--
	}
	for i >= 0 && !os.IsPathSeparator(path[i]) {
		i--
	}
	if i < 0 {
		i = 0
	}

	// If there is a parent directory, and it is not the volume name,
	// recurse to ensure parent directory exists.
	if parent := path[:i]; len(parent) > len(filepath.VolumeName(path)) {
		err = mkdirAllAndChown(parent, perm, uid, gid)
		if err != nil {
			return err
		}
	}

	// Parent now exists; invoke Mkdir and use its result.
	err = os.Mkdir(path, perm)
	if err != nil {
		// Handle arguments like "foo/." by
		// double-checking that directory doesn't exist.
		dir, err1 := os.Lstat(path)
		if err1 == nil && dir.IsDir() {
			return nil
		}
		return err
	}

	err = os.Chown(path, uid, gid)
	if err != nil {
		return err
	}

	return nil
}
