package v1alpha1

type NotificationConfigSpec struct {
	Providers ProvidersConfig `json:"providers"`
}

type ProvidersConfig struct {
	Slack SlackProviderConfig `json:"slack,omitempty"`
}

type SlackWebhookURLSecretRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type SlackProviderConfig struct {
	Enabled             bool                     `json:"enabled"`
	WebhookURLSecretRef SlackWebhookURLSecretRef `json:"webhookURLSecretRef"`
}
