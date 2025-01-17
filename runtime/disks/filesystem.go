package disks

import (
	"context"
	"os/exec"
)

func MkfsEXT4(ctx context.Context, dev string) error {
	return exec.CommandContext(ctx, "mkfs.ext4", "-F", dev).Run()
}
