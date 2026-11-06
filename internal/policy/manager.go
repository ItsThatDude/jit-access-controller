package policy

import (
	"context"
	"slices"
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
	slices.SortFunc(snapshot, func(a, b common.AccessPolicyObject) int {
		aPolicy := a.GetPolicy()
		bPolicy := b.GetPolicy()

		if aPolicy.Priority != bPolicy.Priority {
			switch {
			case aPolicy.Priority > bPolicy.Priority:
				return -1
			case aPolicy.Priority < bPolicy.Priority:
				return 1
			}
		}

		switch {
		case a.GetName() < b.GetName():
			return -1
		case a.GetName() > b.GetName():
			return 1
		default:
			return 0
		}
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
