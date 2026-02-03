package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/nikoksr/notify"
)

type Notifier interface {
	SendApprovalRequest(
		ctx context.Context,
		//msg ApprovalMessage,
		msg ApprovalMessage,
	) error
}

type ApprovalMessage struct {
	RequestName string
	Subject     string
	Resource    string
	ApproveCmd  string
	DenyCmd     string
	ExpiresAt   time.Time
}

type NotifyNotifier struct {
	notify *notify.Notify
}

func NewNotifier(notify *notify.Notify) *NotifyNotifier {
	return &NotifyNotifier{
		notify: notify,
	}
}

func (n *NotifyNotifier) SendApprovalRequest(
	ctx context.Context,
	msg ApprovalMessage,
) error {

	body := fmt.Sprintf(
		"JIT access request\n\n"+
			"User: %s\n"+
			"Resource: %s\n"+
			"Approve: %s\n"+
			"Deny: %s\n"+
			"Expires: %s\n",
		msg.Subject,
		msg.Resource,
		msg.ApproveCmd,
		msg.DenyCmd,
		msg.ExpiresAt.Format(time.RFC3339),
	)

	return n.notify.Send(
		ctx,
		"JIT Access Approval Required",
		body,
	)
}
