package notifications

import (
	"sync"

	"github.com/itsthatdude/jit-access-controller/internal/common"
)

type NotificationManager struct {
	mu     sync.RWMutex
	router *NotificationRouter
}

func (m *NotificationManager) Resolve(
	req *common.AccessRequestObject,
	policy common.AccessPolicyObject,
) Notifier {
	m.mu.RLock()
	defer m.mu.RUnlock()

	router := m.router

	if router == nil {
		return nil
	}

	route := resolveConfigFromPolicy(policy)
	return router.Resolve(route)
}

func (m *NotificationManager) Update(
	key string,
	notifier Notifier,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.router.Update(key, notifier)
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{}
}

func resolveConfigFromPolicy(
	policy common.AccessPolicyObject,
) string {

	var pol = policy.GetPolicy()

	if pol.NotificationConfig != "" {
		return pol.NotificationConfig
	}

	return ""
}
