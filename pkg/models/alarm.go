package models

// AlarmListResponse wraps the plural GET /alarms response
type AlarmListResponse struct {
	Result struct {
		Alarms []Alarm `json:"alarms"`
	} `json:"result"`
}

// Alarm represents a single alarm entry from the list/history endpoint.
type Alarm struct {
	ID          string `json:"id"`
	Name        string `json:"name"`                       
	State       string `json:"state"`                      
	TriggerTime string `json:"timeOfMostRecentActivation"` 
}

// AlarmUpdatePayload is used for PUT /alarm
// UPDATED: Flat structure to match API requirement (no nested "alarm" object)
type AlarmUpdatePayload struct {
	Session string `json:"session"`
	ID      string `json:"id"`
	Action  string `json:"action"`         // API expects "action" (e.g. ACKNOWLEDGE)
	Note    string `json:"note,omitempty"` // Optional
}
