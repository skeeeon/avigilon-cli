package models

// TriggerOutputPayload is the body for PUT /camera/commands/trigger-digital-output
type TriggerOutputPayload struct {
	Session   string `json:"session"`
	ID        string `json:"id,omitempty"`       // Camera ID (triggers all linked outputs)
	EntityID  string `json:"entityId,omitempty"` // Specific Digital Output ID
	IsToggle  bool   `json:"isToggle"`           // Usually true for momentary switches
}
