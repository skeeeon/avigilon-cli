package models

// ManualRecordingPayload is the body for POST /camera/record/manual
type ManualRecordingPayload struct {
	Session        string   `json:"session"`
	CameraIDs      []string `json:"cameraIds"`
	Action         string   `json:"action"`                   // "START" or "STOP"
	MaxDurationSec *int     `json:"maxDurationSec,omitempty"` // Pointer allows nil to omit field
}
