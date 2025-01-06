package files

import (
	"os"
	"path"

	"github.com/valyentdev/ravel/core/errdefs"
)

func validateWriteFilePath(p string) error {
	if p == "" {
		return errdefs.NewInvalidArgument("path must be provided")
	}
	dir, err := os.Stat(path.Dir(p))
	if err != nil {
		if os.IsNotExist(err) {
			return errdefs.NewNotFound("directory does not exist")
		}
		return err
	}

	if !dir.IsDir() {
		return errdefs.NewInvalidArgument("parent is not a directory")
	}

	pathStat, err := os.Stat(p)
	if err == nil {
		if pathStat.IsDir() {
			return errdefs.NewInvalidArgument("path is a directory")
		}

		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	return nil
}
