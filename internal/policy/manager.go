package policy

import (
	"context"
	"sort"
	"sync"

	accessv1alpha1 "github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	common "github.com/itsthatdude/jit-access-controller/internal/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// LoadClusterPolicies lists all ClusterAccessPolicy resources and updates the
// provided PolicyManager with their snapshots. This is useful on startup so
// reconcilers and webhooks have an initial view before informer events arrive.
func LoadClusterPolicies(ctx context.Context, c client.Client, manager *PolicyManager) error {
	var list accessv1alpha1.ClusterAccessPolicyList
	if err := c.List(ctx, &list); err != nil {
		return err
	}

	objs := make([]common.AccessPolicyObject, 0, len(list.Items))
	for i := range list.Items {
		objs = append(objs, &list.Items[i])
	}

	manager.Update(objs)
	return nil
}

// LoadNamespacedPolicies lists all AccessPolicy resources and updates the
// provided PolicyManager. It behaves the same as LoadClusterPolicies but
// targets the namespaced variant of the API.
func LoadNamespacedPolicies(ctx context.Context, c client.Client, manager *PolicyManager) error {
	var list accessv1alpha1.AccessPolicyList
	if err := c.List(ctx, &list); err != nil {
		return err
	}

	objs := make([]common.AccessPolicyObject, 0, len(list.Items))
	for i := range list.Items {
		objs = append(objs, &list.Items[i])
	}

	manager.Update(objs)
	return nil
}
