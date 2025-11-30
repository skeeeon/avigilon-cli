package client

import (
	"fmt"
	"avigilon-cli/pkg/models"
)

func (c *AvigilonClient) GetCameras() ([]models.Camera, error) {
	var respData models.CameraListResponse

	resp, err := c.HTTP.R().
		SetQueryParam("verbosity", "HIGH").
		SetResult(&respData).
		Get("/cameras")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get cameras: %s", resp.String())
	}

	// Navigate the nested structure to get the slice
	return respData.Result.Cameras, nil
}

// TriggerManualRecording starts (or stops) recording on specific cameras.
// action: "START" or "STOP"
// duration: Seconds to record (ignored if action is STOP)
// TriggerManualRecording starts (or stops) recording on specific cameras.
func (c *AvigilonClient) TriggerManualRecording(sessionID string, cameraIDs []string, action string, duration int) error {
	// Logic to handle API requirement: maxDurationSec must NOT be present for STOP
	var durationPtr *int
	if action == "START" {
		durationPtr = &duration
	}

	payload := models.ManualRecordingPayload{
		Session:        sessionID,
		CameraIDs:      cameraIDs,
		Action:         action,
		MaxDurationSec: durationPtr, // Will be nil if STOP, omitting the field
	}

	// Page 31: POST /camera/record/manual
	resp, err := c.HTTP.R().
		SetBody(payload).
		Post("/camera/record/manual")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to trigger recording: %s", resp.String())
	}

	return nil
}
