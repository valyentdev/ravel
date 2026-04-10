package disks

import (
	"testing"
	"time"
)

// mockDevicePool implements DevicePool for testing
type mockDevicePool struct {
	devices   map[string]string
	snapshots map[string]string
}

func newMockDevicePool() *mockDevicePool {
	return &mockDevicePool{
		devices:   make(map[string]string),
		snapshots: make(map[string]string),
	}
}

func (m *mockDevicePool) CreateDevice(id string, sizeMB uint64) (string, error) {
	path := "/dev/mock/" + id
	m.devices[id] = path
	return path, nil
}

func (m *mockDevicePool) DeleteDevice(id string) error {
	delete(m.devices, id)
	return nil
}

func (m *mockDevicePool) Snapshot(id, snapshot string) error {
	m.snapshots[snapshot] = id
	return nil
}

func (m *mockDevicePool) DeleteSnapshot(id, snapshot string) error {
	delete(m.snapshots, snapshot)
	return nil
}

// mockStore implements Store for testing
type mockStore struct {
	disks map[string]*Disk
}

func newMockStore() *mockStore {
	return &mockStore{
		disks: make(map[string]*Disk),
	}
}

func (m *mockStore) BeginDiskTX(writable bool) (DiskTX, error) {
	return &mockDiskTX{
		store:    m,
		writable: writable,
	}, nil
}

type mockDiskTX struct {
	store    *mockStore
	writable bool
}

func (m *mockDiskTX) Commit() error {
	return nil
}

func (m *mockDiskTX) Rollback() error {
	return nil
}

func (m *mockDiskTX) ListDisks() ([]Disk, error) {
	disks := make([]Disk, 0, len(m.store.disks))
	for _, d := range m.store.disks {
		disks = append(disks, *d)
	}
	return disks, nil
}

func (m *mockDiskTX) GetDisk(id string) (*Disk, error) {
	d, ok := m.store.disks[id]
	if !ok {
		return nil, &diskNotFoundError{id: id}
	}
	return d, nil
}

func (m *mockDiskTX) PutDisk(disk *Disk) error {
	m.store.disks[disk.Id] = disk
	return nil
}

func (m *mockDiskTX) DeleteDisk(id string) error {
	delete(m.store.disks, id)
	return nil
}

type diskNotFoundError struct {
	id string
}

func (e *diskNotFoundError) Error() string {
	return "disk not found: " + e.id
}

// Skip MkfsEXT4 in tests - it requires actual disk creation
func init() {
	// In real tests, we'd mock this or skip it
}

func TestServiceGetDisk(t *testing.T) {
	store := newMockStore()
	pool := newMockDevicePool()
	svc := NewService(store, pool)

	// Create a test disk directly in store
	testDisk := &Disk{
		Id:        "test-disk",
		SizeMB:    1024,
		CreatedAt: time.Now(),
		Path:      "/dev/mock/test-disk",
	}
	store.disks["test-disk"] = testDisk

	// Test GetDisk
	disk, err := svc.GetDisk("test-disk")
	if err != nil {
		t.Fatalf("GetDisk() error = %v", err)
	}
	if disk.Id != "test-disk" {
		t.Errorf("GetDisk() id = %s, want test-disk", disk.Id)
	}
	if disk.SizeMB != 1024 {
		t.Errorf("GetDisk() size = %d, want 1024", disk.SizeMB)
	}
}

func TestServiceListDisks(t *testing.T) {
	store := newMockStore()
	pool := newMockDevicePool()
	svc := NewService(store, pool)

	// Add some disks
	store.disks["disk1"] = &Disk{Id: "disk1", SizeMB: 1024, CreatedAt: time.Now()}
	store.disks["disk2"] = &Disk{Id: "disk2", SizeMB: 2048, CreatedAt: time.Now()}

	disks, err := svc.ListDisks()
	if err != nil {
		t.Fatalf("ListDisks() error = %v", err)
	}
	if len(disks) != 2 {
		t.Errorf("ListDisks() count = %d, want 2", len(disks))
	}
}

func TestServiceGetDisks(t *testing.T) {
	store := newMockStore()
	pool := newMockDevicePool()
	svc := NewService(store, pool)

	// Add test disks
	store.disks["disk1"] = &Disk{Id: "disk1", SizeMB: 1024, CreatedAt: time.Now()}
	store.disks["disk2"] = &Disk{Id: "disk2", SizeMB: 2048, CreatedAt: time.Now()}

	disks, err := svc.GetDisks("disk1", "disk2")
	if err != nil {
		t.Fatalf("GetDisks() error = %v", err)
	}
	if len(disks) != 2 {
		t.Errorf("GetDisks() count = %d, want 2", len(disks))
	}
}

func TestServiceAttachDetachInstance(t *testing.T) {
	store := newMockStore()
	pool := newMockDevicePool()
	svc := NewService(store, pool)

	// Create test disk
	store.disks["disk1"] = &Disk{Id: "disk1", SizeMB: 1024, CreatedAt: time.Now()}

	// Test Attach
	err := svc.AttachInstance("instance1", "disk1")
	if err != nil {
		t.Fatalf("AttachInstance() error = %v", err)
	}

	disk, _ := svc.GetDisk("disk1")
	if disk.AttachedInstance != "instance1" {
		t.Errorf("AttachedInstance = %s, want instance1", disk.AttachedInstance)
	}

	// Test Detach
	err = svc.DetachInstance("disk1")
	if err != nil {
		t.Fatalf("DetachInstance() error = %v", err)
	}

	disk, _ = svc.GetDisk("disk1")
	if disk.AttachedInstance != "" {
		t.Errorf("AttachedInstance = %s, want empty", disk.AttachedInstance)
	}
}

func TestServiceAttachAlreadyAttached(t *testing.T) {
	store := newMockStore()
	pool := newMockDevicePool()
	svc := NewService(store, pool)

	// Create disk already attached
	store.disks["disk1"] = &Disk{
		Id:               "disk1",
		SizeMB:           1024,
		CreatedAt:        time.Now(),
		AttachedInstance: "instance1",
	}

	// Try to attach to another instance
	err := svc.AttachInstance("instance2", "disk1")
	if err == nil {
		t.Error("AttachInstance() expected error for already attached disk")
	}
}

func TestDiskConfigGetDisks(t *testing.T) {
	// This would test the InstanceConfig.GetDisks() method
	// which is in core/instance/instance.go
	// We can add this test there
}
