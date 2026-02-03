package notifications

type NotificationRouter struct {
	defaultRoute string
	byRoute      map[string]Notifier
}

func (r *NotificationRouter) Update(
	key string,
	notifier Notifier,
) {
	r.byRoute[key] = notifier
}

func (r *NotificationRouter) Resolve(
	key string,
) Notifier {

	if key != "" {
		if n, ok := r.byRoute[key]; ok {
			return n
		}
	}

	return nil
}

func NewNotificationRouter(
	defaultRoute string,
	routes map[string]Notifier,
) *NotificationRouter {

	return &NotificationRouter{
		defaultRoute: defaultRoute,
		byRoute:      routes,
	}
}
