//go:build !release
// +build !release

package allocrunner

import (
	"sync"
	"testing"

	"github.com/hashicorp/nomad/client/lib/cgutil"

	"github.com/hashicorp/nomad/client/allocwatcher"
	clientconfig "github.com/hashicorp/nomad/client/config"
	"github.com/hashicorp/nomad/client/consul"
	"github.com/hashicorp/nomad/client/devicemanager"
	"github.com/hashicorp/nomad/client/pluginmanager/drivermanager"
	"github.com/hashicorp/nomad/client/state"
	"github.com/hashicorp/nomad/client/vaultclient"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/stretchr/testify/require"
)

// MockStateUpdater implements the AllocStateHandler interface and records
// alloc updates.
type MockStateUpdater struct {
	Updates []*structs.Allocation
	mu      sync.Mutex
}

// AllocStateUpdated implements the AllocStateHandler interface and records an
// alloc update.
func (m *MockStateUpdater) AllocStateUpdated(alloc *structs.Allocation) {
	m.mu.Lock()
	m.Updates = append(m.Updates, alloc)
	m.mu.Unlock()
}

// PutAllocation satisfies the AllocStateHandler interface.
func (m *MockStateUpdater) PutAllocation(alloc *structs.Allocation) (err error) {
	return
}

// Last returns a copy of the last alloc (or nil) update. Safe for concurrent
// access with updates.
func (m *MockStateUpdater) Last() *structs.Allocation {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := len(m.Updates)
	if n == 0 {
		return nil
	}
	return m.Updates[n-1].Copy()
}

// Reset resets the recorded alloc updates.
func (m *MockStateUpdater) Reset() {
	m.mu.Lock()
	m.Updates = nil
	m.mu.Unlock()
}

func testAllocRunnerConfig(t *testing.T, alloc *structs.Allocation) (*Config, func()) {
	clientConf, cleanup := clientconfig.TestClientConfig(t)
	conf := &Config{
		// Copy the alloc in case the caller edits and reuses it
		Alloc:              alloc.Copy(),
		Logger:             clientConf.Logger,
		ClientConfig:       clientConf,
		StateDB:            state.NoopDB{},
		Consul:             consul.NewMockConsulServiceClient(t, clientConf.Logger),
		ConsulSI:           consul.NewMockServiceIdentitiesClient(),
		Vault:              vaultclient.NewMockVaultClient(),
		StateUpdater:       &MockStateUpdater{},
		PrevAllocWatcher:   allocwatcher.NoopPrevAlloc{},
		PrevAllocMigrator:  allocwatcher.NoopPrevAlloc{},
		DeviceManager:      devicemanager.NoopMockManager(),
		DriverManager:      drivermanager.TestDriverManager(t),
		CpusetManager:      cgutil.NoopCpusetManager(),
		ServersContactedCh: make(chan struct{}),
	}
	return conf, cleanup
}

func TestAllocRunnerFromAlloc(t *testing.T, alloc *structs.Allocation) (*allocRunner, func()) {
	t.Helper()
	cfg, cleanup := testAllocRunnerConfig(t, alloc)
	ar, err := NewAllocRunner(cfg)
	if err != nil {
		require.NoError(t, err, "Failed to setup AllocRunner")
	}

	return ar, cleanup
}

// AllocFailer provides access to the concrete allocRunner instance for test code.
type AllocFailer struct {
	Runner *allocRunner
}

// FailTask allows failing a task from test code.
func (af *AllocFailer) FailTask(taskName, taskEvent string) error {
	if taskEvent == "" {
		taskEvent = structs.TaskDriverFailure
	}

	event := structs.NewTaskEvent(taskEvent).SetFailsTask()
	for taskKey, taskRunner := range af.Runner.tasks {
		if taskName == "" || taskName == taskKey {
			taskRunner.AppendEvent(event)
		}
	}

	// Calculate alloc state to get the final state with the new events.
	// Cannot rely on AllocStates as it won't recompute TaskStates once they are set.
	states := make(map[string]*structs.TaskState, len(af.Runner.tasks))
	for name, tr := range af.Runner.tasks {
		taskState := tr.TaskState()
		taskState.State = structs.TaskStateDead
		states[name] = taskState
	}

	// Build the client allocation
	alloc := af.Runner.clientAlloc(states)

	// Update the client state store.
	err := af.Runner.stateUpdater.PutAllocation(alloc)
	if err != nil {
		return err
	}

	// Update the server.
	af.Runner.stateUpdater.AllocStateUpdated(alloc)

	// Broadcast client alloc to listeners.
	err = af.Runner.allocBroadcaster.Send(alloc)

	return err
}
