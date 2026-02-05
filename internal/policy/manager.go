package policy

import (
	"sort"
	"sync"

	common "github.com/itsthatdude/jit-access-controller/internal/common"
)

type PolicyManager struct {
	mu       sync.RWMutex
	policies []common.AccessPolicyObject
}

func NewPolicyManager() *PolicyManager {
	return &PolicyManager{}
}

func (m *PolicyManager) Update(
	policies []common.AccessPolicyObject,
) {
	// Defensive copy
	snapshot := make([]common.AccessPolicyObject, len(policies))
	copy(snapshot, policies)

	// Sort by priority DESC, name ASC (deterministic)
	sort.Slice(snapshot, func(i, j int) bool {
		var name = snapshot[i].GetName()
		var policy = snapshot[i].GetPolicy()

		var lastName = snapshot[j].GetName()
		var lastPolicy = snapshot[j].GetPolicy()

		if policy.Priority != lastPolicy.Priority {
			return policy.Priority > lastPolicy.Priority
		}

		return name < lastName
	})

	m.mu.Lock()
	defer m.mu.Unlock()

	m.policies = snapshot
}

func (m *PolicyManager) GetSnapshot() []common.AccessPolicyObject {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := make([]common.AccessPolicyObject, len(m.policies))
	copy(snapshot, m.policies)
	return snapshot
}
