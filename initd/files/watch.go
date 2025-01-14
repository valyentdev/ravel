package files

import (
	"context"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/api/initd"
)

func (s *Service) WatchDir(ctx context.Context, path string) (<-chan initd.WatchFSEvent, error) {
	dir, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errdefs.NewNotFound("directory does not exist")
		}
		return nil, err
	}

	if !dir.IsDir() {
		return nil, errdefs.NewInvalidArgument("path is not a directory")
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = w.Add(path)
	if err != nil {
		return nil, err
	}

	events := make(chan initd.WatchFSEvent)

	go func() {
		defer w.Close()
		defer close(events)
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-w.Events:
				if event.String() == "CHMOD" {
					continue // ignore CHMOD only events
				}
				select { // non-blocking send in case the context is done
				case events <- initd.WatchFSEvent{
					Path:   event.Name,
					Create: fsnotify.Create.Has(event.Op),
					Write:  fsnotify.Write.Has(event.Op),
					Remove: fsnotify.Remove.Has(event.Op),
					Rename: fsnotify.Rename.Has(event.Op),
				}:
				case <-ctx.Done():
					return
				}
			}
		}

	}()

	return events, nil
}
