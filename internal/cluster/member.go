package cluster

import (
	"sync"
)

// Member describes a node in the twodev cluster.
type Member struct {
	Name string
	IP   string
	Port int
}

// Registry tracks cluster members in memory.
type Registry struct {
	mu      sync.RWMutex
	local   Member
	members []Member
}

// NewRegistry creates a cluster registry with the local node.
func NewRegistry(local Member) *Registry {
	return &Registry{local: local, members: []Member{local}}
}

// Local returns this node.
func (r *Registry) Local() Member {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.local
}

// Members returns known cluster members.
func (r *Registry) Members() []Member {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Member, len(r.members))
	copy(out, r.members)
	return out
}

// Upsert adds or replaces a member.
func (r *Registry) Upsert(member Member) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, existing := range r.members {
		if existing.Name == member.Name {
			r.members[i] = member
			return
		}
	}
	r.members = append(r.members, member)
}