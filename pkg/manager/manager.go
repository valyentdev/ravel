package manager

import (
	"github.com/nats-io/nats.go"
	corro "github.com/valyentdev/ravel/internal/cluster"
	"github.com/valyentdev/ravel/internal/placement"
	"github.com/valyentdev/ravel/pkg/helper/corroclient"
)

type VCPUKindTemplate struct {
	VCpuFrequency   int `json:"vcpu_frequency"` // MHz
	MinVCpus        int `json:"min_vcpus"`
	MaxVCpus        int `json:"max_vcpus"`
	MinMemoryByVCPU int `json:"min_memory_by_vcpu"`
	MaxMemoryByVCPU int `json:"max_memory_by_vcpu"`
}

var defaultCPUSpecs = map[string]VCPUKindTemplate{
	"eco": {
		VCpuFrequency:   240,
		MinVCpus:        1,
		MaxVCpus:        4,
		MinMemoryByVCPU: 256,
		MaxMemoryByVCPU: 2048,
	},
	"performance": {
		VCpuFrequency:   2400,
		MinVCpus:        1,
		MaxVCpus:        4,
		MinMemoryByVCPU: 512,
		MaxMemoryByVCPU: 4096,
	},
}

func IsResourcesConfigValid(vcpus int, memory int, template VCPUKindTemplate) bool {
	isValid := memory%256 == 0 && (vcpus == 1 || vcpus%2 == 0) &&
		vcpus >= template.MinVCpus &&
		vcpus <= template.MaxVCpus &&
		memory/vcpus >= template.MinMemoryByVCPU &&
		memory/vcpus <= template.MaxMemoryByVCPU

	return isValid
}

type ManagerConfig struct {
	CorrosionConfig corroclient.Config
	NatsURl         string
	NatsOptions     []nats.Option
}

func New(config ManagerConfig) (*Manager, error) {
	nc, err := nats.Connect(config.NatsURl, config.NatsOptions...)
	if err != nil {
		return nil, err
	}

	clusterState, err := corro.Connect(config.CorrosionConfig)
	if err != nil {
		return nil, err
	}

	broker := placement.NewBroker(nc)

	return &Manager{
		nc:           nc,
		clusterState: clusterState,
		broker:       broker,
	}, nil
}

type Manager struct {
	clusterState *corro.ClusterState
	nc           *nats.Conn
	broker       *placement.Broker
}

func (m *Manager) Close() {
	m.nc.Close()
}
