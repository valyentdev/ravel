package validation

import (
	"errors"
	"regexp"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/internal/signals"
)

var idRegexp = regexp.MustCompile("^[A-Za-z0-9_-]+$")

func ValidateInstanceId(id string) error {
	if id == "" {
		return errors.New("instance id is required")
	}

	if len(id) > 64 {
		return errors.New("instance id is too long")
	}

	if !idRegexp.MatchString(id) {
		return errors.New("instance id contains invalid characters")
	}
	return nil
}

func getPath(basePath []string) string {
	if len(basePath) == 0 {
		return "body"
	}

	return basePath[0]
}

func ValidateStopConfig(config *api.StopConfig, basePath ...string) []*errdefs.ErrorDetail {
	path := getPath(basePath)

	var errDetails []*errdefs.ErrorDetail
	if config != nil {
		if config.Signal != nil {
			_, ok := signals.FromString(*config.Signal)
			if !ok {
				errDetails = append(errDetails, &errdefs.ErrorDetail{
					Location: joinFieldPath(path, "signal"),
					Message:  "invalid signal in stop config",
					Value:    *config.Signal,
				})
			}
		}

		if config.Timeout != nil {
			timeout := *config.Timeout
			if timeout < 0 || timeout > api.MaxStopTimeout {
				errDetails = append(errDetails, &errdefs.ErrorDetail{
					Location: joinFieldPath(path, "timeout"),
					Message:  "invalid timeout",
					Value:    *config.Timeout,
				})
			}
		}
	}

	return errDetails
}

func validateInstanceGuestConfig(config *instance.InstanceGuestConfig, errPath ...string) []*errdefs.ErrorDetail {
	path := getPath(errPath)
	var errDetails []*errdefs.ErrorDetail
	if config.CpusMHz <= 0 {
		errDetails = append(errDetails, &errdefs.ErrorDetail{
			Location: joinFieldPath(path, "cpus"),
			Message:  "cpus must be greater than 0",
			Value:    config.CpusMHz,
		})
	}

	if config.MemoryMB <= 0 {
		errDetails = append(errDetails, &errdefs.ErrorDetail{
			Location: joinFieldPath(path, "memory_mb"),
			Message:  "memory must be greater than 0",
			Value:    config.MemoryMB,
		})
	}

	if config.MemoryMB%4 != 0 {
		errDetails = append(errDetails, &errdefs.ErrorDetail{
			Location: joinFieldPath(path, "memory_mb"),
			Message:  "memory must be a multiple of 4",
			Value:    config.MemoryMB,
		})
	}

	if config.VCpus <= 0 {
		errDetails = append(errDetails, &errdefs.ErrorDetail{
			Location: joinFieldPath(path, "vcpus"),
			Message:  "vcpus must be greater than 0",
			Value:    config.VCpus,
		})
	}

	return errDetails
}

func ValidateInstanceConfig(config *instance.InstanceConfig, errpath ...string) []*errdefs.ErrorDetail {
	path := getPath(errpath)
	var errDetails []*errdefs.ErrorDetail
	if config.Stop != nil {
		errDetails = append(errDetails, ValidateStopConfig(config.Stop, path)...)
	}

	errDetails = append(errDetails, validateInstanceGuestConfig(&config.Guest, path)...)

	if config.Image == "" {
		errDetails = append(errDetails, &errdefs.ErrorDetail{
			Location: joinFieldPath(path, "image"),
			Message:  "image is required",
		})
	}

	return errDetails
}
