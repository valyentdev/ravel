package build

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/internal/id"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/session"
	"github.com/tonistiigi/fsutil"
	"golang.org/x/sync/errgroup"
)

// BuildOptions contains all options for a build
type BuildOptions struct {
	Namespace  string
	ImageName  string
	Tag        string
	Registry   string
	Dockerfile string
	BuildArgs  map[string]string
	Target     string
	NoCache    bool
	Context    io.Reader // tar.gz stream of build context
}

// BuildResult contains the result of a successful build
type BuildResult struct {
	Digest string
}

// StartBuild initiates a new build and returns immediately
// The build runs asynchronously and its status can be queried via GetBuild
func (s *Service) StartBuild(ctx context.Context, opts BuildOptions) (*api.Build, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("build service is not enabled")
	}

	// Generate build ID
	buildId := "build_" + id.Generate()

	// Set defaults
	if opts.Tag == "" {
		opts.Tag = "latest"
	}
	if opts.Dockerfile == "" {
		opts.Dockerfile = "Dockerfile"
	}

	// Construct full image reference
	fullImage := fmt.Sprintf("%s/%s:%s", opts.Registry, opts.ImageName, opts.Tag)

	// Create build record
	build := &api.Build{
		Id:        buildId,
		Namespace: opts.Namespace,
		ImageName: opts.ImageName,
		Tag:       opts.Tag,
		Registry:  opts.Registry,
		FullImage: fullImage,
		Status:    api.BuildStatusPending,
		CreatedAt: time.Now(),
	}

	// Save to store
	if err := s.store.CreateBuild(ctx, build); err != nil {
		return nil, fmt.Errorf("failed to create build record: %w", err)
	}

	// Extract context to temp directory
	contextDir, err := s.extractContext(opts.Context)
	if err != nil {
		build.Status = api.BuildStatusFailed
		build.Error = fmt.Sprintf("failed to extract context: %v", err)
		s.store.UpdateBuild(ctx, build)
		return build, nil
	}

	// Create log file
	logFile := filepath.Join(os.TempDir(), "ravel-builds", buildId+".log")
	os.MkdirAll(filepath.Dir(logFile), 0755)

	// Create cancellable context for the build
	buildCtx, cancel := context.WithCancel(context.Background())

	// Track build state
	state := &buildState{
		build:   build,
		cancel:  cancel,
		logFile: logFile,
	}

	s.buildsLock.Lock()
	s.builds[buildId] = state
	s.buildsLock.Unlock()

	// Start build in background
	go s.runBuild(buildCtx, state, contextDir, opts)

	return build, nil
}

// runBuild executes the actual build process
func (s *Service) runBuild(ctx context.Context, state *buildState, contextDir string, opts BuildOptions) {
	build := state.build
	startTime := time.Now()

	defer func() {
		// Cleanup context directory
		os.RemoveAll(contextDir)

		// Remove from active builds
		s.buildsLock.Lock()
		delete(s.builds, build.Id)
		s.buildsLock.Unlock()

		// Release semaphore
		<-s.sem
	}()

	// Acquire semaphore (wait for slot)
	select {
	case s.sem <- struct{}{}:
	case <-ctx.Done():
		build.Status = api.BuildStatusFailed
		build.Error = "build cancelled while waiting"
		s.store.UpdateBuild(context.Background(), build)
		return
	}

	// Update status to building
	build.Status = api.BuildStatusBuilding
	s.store.UpdateBuild(context.Background(), build)

	slog.Info("Starting build", "build_id", build.Id, "image", build.FullImage)

	// Open log file for writing
	logFile, err := os.Create(state.logFile)
	if err != nil {
		slog.Error("Failed to create log file", "error", err)
	}
	defer logFile.Close()

	// Build the image
	digest, err := s.executeBuild(ctx, contextDir, opts, build.FullImage, logFile)
	if err != nil {
		build.Status = api.BuildStatusFailed
		build.Error = err.Error()
		now := time.Now()
		build.CompletedAt = &now
		build.DurationMs = time.Since(startTime).Milliseconds()
		s.store.UpdateBuild(context.Background(), build)
		slog.Error("Build failed", "build_id", build.Id, "error", err)
		return
	}

	// Success
	build.Status = api.BuildStatusCompleted
	build.Digest = digest
	now := time.Now()
	build.CompletedAt = &now
	build.DurationMs = time.Since(startTime).Milliseconds()
	s.store.UpdateBuild(context.Background(), build)

	slog.Info("Build completed", "build_id", build.Id, "digest", digest, "duration_ms", build.DurationMs)
}

// executeBuild performs the actual BuildKit solve operation
func (s *Service) executeBuild(ctx context.Context, contextDir string, opts BuildOptions, fullImage string, logWriter io.Writer) (string, error) {
	// Create local filesystem for context
	contextFS, err := fsutil.NewFS(contextDir)
	if err != nil {
		return "", fmt.Errorf("failed to create context fs: %w", err)
	}

	// Create solve options
	solveOpt := client.SolveOpt{
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			"filename": opts.Dockerfile,
		},
		LocalMounts: map[string]fsutil.FS{
			"dockerfile": contextFS,
			"context":    contextFS,
		},
		Exports: []client.ExportEntry{
			{
				Type: client.ExporterImage,
				Attrs: map[string]string{
					"name": fullImage,
					"push": "true",
				},
			},
		},
	}

	// Add build args
	for k, v := range opts.BuildArgs {
		solveOpt.FrontendAttrs["build-arg:"+k] = v
	}

	// Add target if specified
	if opts.Target != "" {
		solveOpt.FrontendAttrs["target"] = opts.Target
	}

	// Add no-cache if specified
	if opts.NoCache {
		solveOpt.FrontendAttrs["no-cache"] = ""
	}

	// Create session for authentication using Ravel's registry config
	solveOpt.Session = []session.Attachable{newAuthProvider(s.registries)}

	// Create progress channel
	ch := make(chan *client.SolveStatus)

	// Run solve with progress logging
	eg, ctx := errgroup.WithContext(ctx)

	var solveRes *client.SolveResponse

	eg.Go(func() error {
		var err error
		solveRes, err = s.client.Solve(ctx, nil, solveOpt, ch)
		return err
	})

	eg.Go(func() error {
		// Write progress to log
		for status := range ch {
			for _, v := range status.Vertexes {
				if v.Name != "" {
					if v.Started != nil {
						fmt.Fprintf(logWriter, "[%s] %s\n", v.Started.Format(time.RFC3339), v.Name)
					} else {
						fmt.Fprintf(logWriter, "%s\n", v.Name)
					}
				}
				if v.Error != "" {
					fmt.Fprintf(logWriter, "[ERROR] %s\n", v.Error)
				}
			}
			for _, l := range status.Logs {
				fmt.Fprintf(logWriter, "%s", l.Data)
			}
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return "", fmt.Errorf("build failed: %w", err)
	}

	// Get digest from response
	digest := ""
	if solveRes != nil && solveRes.ExporterResponse != nil {
		if d, ok := solveRes.ExporterResponse["containerimage.digest"]; ok {
			digest = d
		}
	}

	if digest == "" {
		digest = fullImage // Fallback to image name
	}

	return digest, nil
}

// extractContext extracts a tar.gz stream to a temporary directory
func (s *Service) extractContext(r io.Reader) (string, error) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "ravel-build-context-")
	if err != nil {
		return "", err
	}

	// Create gzip reader
	gzr, err := gzip.NewReader(r)
	if err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Extract files
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			os.RemoveAll(dir)
			return "", fmt.Errorf("failed to read tar: %w", err)
		}

		target := filepath.Join(dir, header.Name)

		// Security check - prevent path traversal
		if !filepath.HasPrefix(target, dir) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				os.RemoveAll(dir)
				return "", err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				os.RemoveAll(dir)
				return "", err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				os.RemoveAll(dir)
				return "", err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				os.RemoveAll(dir)
				return "", err
			}
			f.Close()
		}
	}

	return dir, nil
}

// getDockerConfigDir returns the path to docker config for auth
func (s *Service) getDockerConfigDir() string {
	// Check for custom config path
	if configDir := os.Getenv("DOCKER_CONFIG"); configDir != "" {
		return configDir
	}
	// Default to ~/.docker
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".docker")
}

// GetBuildLogs returns the log file path for a build
func (s *Service) GetBuildLogs(buildId string) (string, error) {
	s.buildsLock.RLock()
	state, ok := s.builds[buildId]
	s.buildsLock.RUnlock()

	if ok && state.logFile != "" {
		return state.logFile, nil
	}

	// Check if log file exists in temp directory
	logFile := filepath.Join(os.TempDir(), "ravel-builds", buildId+".log")
	if _, err := os.Stat(logFile); err == nil {
		return logFile, nil
	}

	return "", fmt.Errorf("logs not found for build %s", buildId)
}
