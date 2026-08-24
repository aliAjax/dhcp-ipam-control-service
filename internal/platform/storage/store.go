package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Network struct {
	ID, Name, CIDR string
	Labels         map[string]string
	CreatedAt      time.Time
}
type Subnet struct {
	ID, NetworkID, CIDR string
	Gateway             string
	DNS                 []string
	LeaseTTL            time.Duration
	State               string
}
type Pool struct {
	ID, SubnetID, Start, End string
	Excluded                 []string
	State                    string
}
type Lease struct {
	ID, PoolID, ClientID, Address, Family, State string
	ExpiresAt, UpdatedAt                         time.Time
	Version                                      uint64
}
type Conflict struct {
	ID, Address, Reason string
	DetectedAt          time.Time
	Resolved            bool
	State               string
}

type Store interface {
	CreateNetwork(context.Context, Network) error
	CreateSubnet(context.Context, Subnet) error
	GetSubnet(context.Context, string) (Subnet, bool)
	UpdateSubnetState(context.Context, string, string) error
	DeleteSubnet(context.Context, string) error
	CreatePool(context.Context, Pool) error
	GetPool(context.Context, string) (Pool, bool)
	UpdatePoolState(context.Context, string, string) error
	DeletePool(context.Context, string) error
	DeletePoolsBySubnet(context.Context, string) error
	PutLease(context.Context, Lease) error
	GetLease(context.Context, string) (Lease, bool)
	ListLeases(context.Context, string, int) ([]Lease, string)
	ReleaseLease(context.Context, string) error
	AddConflict(context.Context, Conflict) error
	UpdateConflictState(context.Context, string, string) error
	ListConflicts(context.Context) []Conflict
}

type Memory struct {
	mu               sync.RWMutex
	networks         map[string]Network
	subnets          map[string]Subnet
	pools            map[string]Pool
	leases           map[string]Lease
	conflicts        map[string]Conflict
	failSubnetCreate error
	failPoolCreate   error
}

func NewMemory() *Memory {
	return &Memory{networks: map[string]Network{}, subnets: map[string]Subnet{}, pools: map[string]Pool{}, leases: map[string]Lease{}, conflicts: map[string]Conflict{}}
}
func (m *Memory) FailNextSubnetCreate(err error) {
	m.mu.Lock()
	m.failSubnetCreate = err
	m.mu.Unlock()
}
func (m *Memory) FailNextPoolCreate(err error) {
	m.mu.Lock()
	m.failPoolCreate = err
	m.mu.Unlock()
}
func (m *Memory) CreateNetwork(_ context.Context, n Network) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.networks {
		if x.CIDR == n.CIDR {
			return fmt.Errorf("network cidr already exists")
		}
	}
	m.networks[n.ID] = n
	return nil
}
func (m *Memory) CreateSubnet(_ context.Context, s Subnet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.networks[s.NetworkID]; !ok {
		return fmt.Errorf("network not found")
	}
	m.subnets[s.ID] = s
	if m.failSubnetCreate != nil {
		err := m.failSubnetCreate
		m.failSubnetCreate = nil
		return err
	}
	return nil
}
func (m *Memory) GetSubnet(_ context.Context, id string) (Subnet, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.subnets[id]
	return s, ok
}
func (m *Memory) UpdateSubnetState(_ context.Context, id, state string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.subnets[id]
	if !ok {
		return fmt.Errorf("subnet not found")
	}
	s.State = state
	return nil
}
func (m *Memory) DeleteSubnet(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subnets, id)
	return nil
}
func (m *Memory) CreatePool(_ context.Context, p Pool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.subnets[p.SubnetID]; !ok {
		return fmt.Errorf("subnet not found")
	}
	m.pools[p.ID] = p
	if m.failPoolCreate != nil {
		err := m.failPoolCreate
		m.failPoolCreate = nil
		return err
	}
	return nil
}
func (m *Memory) GetPool(_ context.Context, id string) (Pool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.pools[id]
	return p, ok
}
func (m *Memory) UpdatePoolState(_ context.Context, id, state string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.pools[id]
	if !ok {
		return fmt.Errorf("pool not found")
	}
	p.State = state
	return nil
}
func (m *Memory) DeletePool(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pools, id)
	return nil
}
func (m *Memory) DeletePoolsBySubnet(_ context.Context, subnetID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, p := range m.pools {
		if p.SubnetID == subnetID {
			delete(m.pools, id)
		}
	}
	return nil
}
func (m *Memory) PutLease(_ context.Context, l Lease) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.leases[l.ID]; ok && old.Version > l.Version {
		return fmt.Errorf("stale lease version")
	}
	m.leases[l.ID] = l
	return nil
}
func (m *Memory) GetLease(_ context.Context, id string) (Lease, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.leases[id]
	return l, ok
}
func (m *Memory) ListLeases(_ context.Context, cursor string, limit int) ([]Lease, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Lease, 0, limit)
	next := ""
	started := cursor == ""
	for id, l := range m.leases {
		if !started {
			if id == cursor {
				started = true
			}
			continue
		}
		if len(out) == limit {
			next = id
			break
		}
		out = append(out, l)
	}
	return out, next
}
func (m *Memory) ReleaseLease(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.leases[id]
	if !ok {
		return fmt.Errorf("lease not found")
	}
	l.State = "released"
	l.UpdatedAt = time.Now()
	l.Version++
	m.leases[id] = l
	return nil
}
func (m *Memory) AddConflict(_ context.Context, c Conflict) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conflicts[c.ID] = c
	return nil
}
func (m *Memory) UpdateConflictState(_ context.Context, id, state string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.conflicts[id]
	if !ok {
		return fmt.Errorf("conflict not found")
	}
	c.State = state
	c.Resolved = state == "resolved"
	return nil
}
func (m *Memory) ListConflicts(_ context.Context) []Conflict {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Conflict, 0, len(m.conflicts))
	for _, c := range m.conflicts {
		out = append(out, c)
	}
	return out
}
func (m *Memory) Expire(now time.Time) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for id, l := range m.leases {
		if l.State == "active" && now.After(l.ExpiresAt) {
			l.State = "expired"
			l.Version++
			l.UpdatedAt = now
			m.leases[id] = l
			n++
		}
	}
	return n
}
