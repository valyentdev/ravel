package resources

import (
	"errors"

	"github.com/c9s/goprocinfo/linux"
)

func GetHostCPUFrequency() (int64, error) {
	infos, err := linux.ReadCPUInfo("/proc/cpuinfo")
	if err != nil {
		return 0, err
	}

	if len(infos.Processors) == 0 {
		return 0, errors.New("no CPU info found")
	}

	return int64(infos.Processors[0].MHz), nil
}
