package models

// CameraListResponse represents the outer wrapper of the API response
type CameraListResponse struct {
	Result struct {
		Cameras []Camera `json:"cameras"`
	} `json:"result"`
}

// Camera represents a single Avigilon camera device
type Camera struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Model           string `json:"model"`
	Serial          string `json:"serial"` // JSON key is "serial", not "serialNumber"
	FirmwareVersion string `json:"firmwareVersion"`
	ConnectionState string `json:"connectionState"`
	IPAddress       string `json:"ipAddress"` // JSON key is singular string "ipAddress"
	Connected       bool   `json:"connected"`
	RecordedData    bool   `json:"recordedData"`
}
