package files

import (
	"context"
	"io"
	"os"
	"path"

	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/api/initd"
)

type Service struct{}

func (s *Service) ListDir(ctx context.Context, dir string) ([]initd.FSEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errdefs.NewNotFound("directory does not exist")
		}
		return nil, err
	}
	var result []initd.FSEntry
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		result = append(result, initd.FSEntry{
			Name:    entry.Name(),
			Path:    path.Join(dir, entry.Name()),
			IsDir:   info.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
		})
	}

	return result, nil
}

const defaultMode = 0755

func (s *Service) Mkdir(ctx context.Context, path string) error {
	err := os.MkdirAll(path, defaultMode)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) OpenFile(ctx context.Context, path string) (*os.File, error) {
	infos, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errdefs.NewNotFound("file does not exist")
		}
		return nil, err
	}

	if infos.IsDir() {
		return nil, errdefs.NewInvalidArgument("path is a directory")
	}

	if !infos.Mode().IsRegular() {
		return nil, errdefs.NewInvalidArgument("path is not a regular file")
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errdefs.NewNotFound("file does not exist")
		}
		return nil, err
	}

	return file, nil
}

// Remove implements initd.FileService.
func (s *Service) Remove(ctx context.Context, path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

func (s *Service) Rename(ctx context.Context, oldPath string, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}

func (s *Service) Stat(ctx context.Context, path string) (*initd.FSEntry, error) {
	entry, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errdefs.NewNotFound("file does not exist")
		}
		return nil, err
	}
	return &initd.FSEntry{
		Name:    entry.Name(),
		Path:    path,
		IsDir:   entry.IsDir(),
		Size:    entry.Size(),
		ModTime: entry.ModTime().Unix(),
	}, nil
}

func (s *Service) WriteFile(ctx context.Context, p string, content io.Reader) error {
	if err := validateWriteFilePath(p); err != nil {
		return err
	}

	err := s.Mkdir(ctx, path.Dir(p))
	if err != nil {
		return err
	}

	file, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, defaultMode)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, content); err != nil {
		return err
	}

	return nil
}
