// MIT License

// Copyright (c) 2020 Milan Pavlik

package logging

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func DefaultFilenameFunc() string {
	return fmt.Sprintf("%s-%s.log", time.Now().UTC().Format(time.RFC3339), RandomHash(3))
}

// Options define configuration options for Writer
type Options struct {
	// Directory defines the directory where log files will be written to.
	// If the directory does not exist, it will be created.
	Directory string

	// MaximumFileSize defines the maximum size of each log file in bytes.
	// When MaximumFileSize == 0, no upper bound will be enforced.
	// No file will be greater than MaximumFileSize. A Write() which would
	// exceed MaximumFileSize will instead cause a new file to be created.
	// If a Write() is attempting to write more bytes than specified by
	// MaximumFileSize, the write will be skipped.
	MaximumFileSize int64

	// MaximumLifetime defines the maximum amount of time a file will
	// be written to before a rotation occurs.
	// When MaximumLifetime == 0, no log rotation will occur.
	MaximumLifetime time.Duration

	// FileNameFunc specifies the name a new file will take.
	// FileNameFunc must ensure collisions in filenames do not occur.
	// Do not rely on timestamps to be unique, high throughput writes
	// may fall on the same timestamp.
	// Eg.
	// 	2020-03-28_15-00-945-<random-hash>.log
	// When FileNameFunc is not specified, DefaultFilenameFunc will be used.
	FileNameFunc func() string

	// FlushAfterEveryWrite specifies whether the writer should flush
	// the buffer after every write.
	FlushAfterEveryWrite bool
}

// RotateWriter is a concurrency-safe writer with file rotation.
type RotateWriter struct {
	logger *log.Logger

	// opts are the configuration options for this Writer
	opts Options

	// f is the currently open file used for appends.
	// Writes to f are only synchronized once Close() is called,
	// or when files are being rotated.
	f *os.File
	// bw is a buffered writer for writing to f
	bw *bufio.Writer
	// bytesWritten is the number of bytes written to f so far,
	// used for size based rotation
	bytesWritten int64
	// ts is the creation timestamp of f,
	// used for time based log rotation
	ts time.Time

	// queue of entries awaiting to be written
	queue chan []byte
	// synchronize write which have started but not been queued up
	pending sync.WaitGroup
	// singal the writer should close
	closing chan struct{}
	// signal the writer has finished writing all queued up entries.
	done chan struct{}
}

func (w *RotateWriter) ReadAll() ([]byte, error) {
	return os.ReadFile(w.f.Name())
}

// Write writes p into the current file, rotating if necessary.
// Write is non-blocking, if the writer's queue is not full.
// Write is blocking otherwise.
func (w *RotateWriter) Write(p []byte) (n int, err error) {
	select {
	case <-w.closing:
		return 0, fmt.Errorf("writer is closed")
	default:
		w.pending.Add(1)
		defer w.pending.Done()
	}

	w.queue <- p

	return len(p), nil
}

// Close closes the writer.
// Any accepted writes will be flushed. Any new writes will be rejected.
// Once Close() exits, files are synchronized to disk.
func (w *RotateWriter) Close() error {
	close(w.closing)
	w.pending.Wait()

	close(w.queue)
	<-w.done

	if w.f != nil {
		if err := w.closeCurrentFile(); err != nil {
			return err
		}
	}

	return nil
}

func (w *RotateWriter) listen() {
	for b := range w.queue {
		if w.f == nil {
			if err := w.rotate(); err != nil {
				w.logger.Println("Failed to create log file", err)
			}
		}

		size := int64(len(b))

		if w.opts.MaximumFileSize != 0 && size > w.opts.MaximumFileSize {
			w.logger.Println("Attempting to write more bytes than allowed by MaximumFileSize. Skipping.")
			continue
		}
		if w.opts.MaximumFileSize != 0 && w.bytesWritten+size > w.opts.MaximumFileSize {
			if err := w.rotate(); err != nil {
				w.logger.Println("Failed to rotate log file", err)
			}
		}

		if w.opts.MaximumLifetime != 0 && time.Now().After(w.ts.Add(w.opts.MaximumLifetime)) {
			if err := w.rotate(); err != nil {
				w.logger.Println("Failed to rotate log file", err)
			}
		}

		if _, err := w.bw.Write(b); err != nil {
			w.logger.Println("Failed to write to file.", err)
		}

		if w.opts.FlushAfterEveryWrite {
			if err := w.flushCurrentFile(); err != nil {
				w.logger.Println("Failed to flush to file.", err)
			}
		}

		w.bytesWritten += size
	}

	close(w.done)
}

func (w *RotateWriter) flushCurrentFile() error {
	if err := w.bw.Flush(); err != nil {
		return fmt.Errorf("failed to flush current log file: %w", err)
	}

	if err := w.f.Sync(); err != nil {
		return fmt.Errorf("failed to sync current log file: %w", err)
	}

	w.bytesWritten = 0

	return nil
}

func (w *RotateWriter) closeCurrentFile() error {
	if err := w.flushCurrentFile(); err != nil {
		return err
	}

	return nil
}

func (w *RotateWriter) rotate() error {
	if w.f != nil {
		if err := w.closeCurrentFile(); err != nil {
			return err
		}
	}

	path := filepath.Join(w.opts.Directory, w.opts.FileNameFunc())
	f, err := newFile(path)
	if err != nil {
		return fmt.Errorf("failed to create new log file at path %v: %w", path, err)
	}

	w.bw = bufio.NewWriter(f)
	w.f = f
	w.bytesWritten = 0
	w.ts = time.Now().UTC()

	return nil
}

// New creates a new concurrency safe Writer which performs log rotation.
func New(logger *log.Logger, opts Options) (*RotateWriter, error) {
	if _, err := os.Stat(opts.Directory); os.IsNotExist(err) {
		if err := os.MkdirAll(opts.Directory, os.ModePerm); err != nil {
			return nil, fmt.Errorf("directory %v does not exist and could not be created", opts.Directory)
		}
	}

	if opts.FileNameFunc == nil {
		opts.FileNameFunc = DefaultFilenameFunc
	}

	w := &RotateWriter{
		logger:  logger,
		opts:    opts,
		queue:   make(chan []byte, 1024),
		closing: make(chan struct{}),
		done:    make(chan struct{}),
	}

	go w.listen()

	return w, nil
}

func newFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
}
