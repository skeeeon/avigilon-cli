package models

// WebhookListResponse wraps the GET /webhooks response
type WebhookListResponse struct {
	Result struct {
		Webhooks []Webhook `json:"webhooks"`
	} `json:"result"`
}

// WebhookPayload represents the body for POST /webhooks
type WebhookPayload struct {
	Session string  `json:"session"` // Added: Required in body per your request
	Webhook Webhook `json:"webhook"`
}

type Webhook struct {
	ID                  string       `json:"id,omitempty"`
	URL                 string       `json:"url"`
	AuthenticationToken string       `json:"authenticationToken"`
	Heartbeat           *Heartbeat   `json:"heartbeat,omitempty"`
	EventTopics         *EventTopics `json:"eventTopics,omitempty"`
}

type Heartbeat struct {
	Enable      bool `json:"enable"`
	FrequencyMs int  `json:"frequencyMs,omitempty"` // Added: Frequency field
}

type EventTopics struct {
	// Changed from 'whitelist' to 'include' based on your specific server requirement
	Include []string `json:"include"` 
}
