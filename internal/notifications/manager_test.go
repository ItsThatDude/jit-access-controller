package notifications

import (
	"context"
)

type fakeNotifier struct{}

func (n *fakeNotifier) SendApprovalRequest(
	ctx context.Context,
	msg ApprovalMessage,
) error {
	return nil
}

/*func TestNotificationManager(t *testing.T) {
	mgr := NewNotificationManager()

	notifier := &fakeNotifier{}

	mgr.Update("default", notifier)

	policy := &accessv1alpha1.AccessPolicy{}
	got := mgr.Resolve(policy)

	if got != notifier {
		t.Fatalf("unexpected notifier")
	}
}*/
