package vm

import (
	"os"
	"path"
)

type Dir string

func (d Dir) String() string {
	return string(d)
}

func (d Dir) InstancesDir() string {
	return path.Join(d.String(), "instances")
}

func (d Dir) InstanceDir(id string) string {
	return path.Join(d.InstancesDir(), id)
}

const DefaultDataDir = Dir("/var/lib/ravel")
const DefaultRunDir = Dir("/var/run/ravel")

func createInstanceDirectories(dataDir Dir, runDir Dir, id string) error {
	if err := os.MkdirAll(dataDir.InstanceDir(id), 0644); err != nil {
		return err
	}

	if err := os.MkdirAll(runDir.InstanceDir(id), 0644); err != nil {
		return err
	}
	return nil
}

func removeInstanceDirectories(dataDir Dir, runDir Dir, id string) error {
	if err := os.RemoveAll(dataDir.InstanceDir(id)); err != nil {
		return err
	}

	if err := os.RemoveAll(runDir.InstanceDir(id)); err != nil {
		return err
	}

	return nil
}
