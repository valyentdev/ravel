package agent

import (
	"context"

	"github.com/valyentdev/ravel/agent/machine/state"
	"github.com/valyentdev/ravel/core/cluster"
)

type stateReporter struct {
	cluster cluster.ClusterState
}

var _ state.StateReporter = (*stateReporter)(nil)

func newStateReporter(cluster cluster.ClusterState) *stateReporter {
	return &stateReporter{cluster: cluster}
}

// DeleteInstanceState implements state.StateReporter.
func (s *stateReporter) DeleteInstanceState(id string) error {
	return nil
}

// UpsertInstanceState implements state.StateReporter.
func (s *stateReporter) UpsertInstanceState(i cluster.MachineInstance) error {
	return s.cluster.UpsertInstance(context.Background(), i)
}

func NewStateReporter(cluster cluster.ClusterState) *stateReporter {
	return &stateReporter{cluster: cluster}
}
