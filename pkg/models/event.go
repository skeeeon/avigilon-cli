package models

type EventListResponse struct {
	Result struct {
		Events []Event `json:"events"`
		Token  string  `json:"token"` // For pagination/continuation
	} `json:"result"`
}

type Event struct {
	ID        string `json:"thisId"`
	Type      string `json:"type"`      // e.g. "DEVICE_MOTION_START", "USER_LOGIN"
	Timestamp string `json:"timestamp"` // ISO 8601
	Server    string `json:"originatingServerName"`
	CameraID  string `json:"cameraId,omitempty"`
	UserName  string `json:"userName,omitempty"`
}
