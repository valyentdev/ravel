package container

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
)

func (r *Runtime) DestroyVM(ctx context.Context, instanceId string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	vm, ok := r.runningVMs[instanceId]
	if !ok {
		return nil
	}

	select {
	case <-vm.waitChan:
	default:
		return errors.New("instance is not exited")
	}
	delete(r.runningVMs, instanceId)
	return nil
}

func (r *Runtime) DestroyInstance(ctx context.Context, instanceId string) error {
	_, ok := r.runningVMs[instanceId]
	if ok {
		return fmt.Errorf("instance %q is still running or vm has not been destroyed", instanceId)
	}
	err := r.fsBuilder.CleanupFilesystems(instanceId)
	if err != nil {
		slog.Error("failed to cleanup filesystems", "error", err)
	}

	err = os.RemoveAll(getInstancePath(instanceId))
	if err != nil {
		slog.Error("failed to remove initrd", "error", err)
	}

	return nil
}
