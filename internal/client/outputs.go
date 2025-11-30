package client

import (
	"fmt"
	"avigilon-cli/pkg/models"
)

// TriggerDigitalOutput activates a digital output.
// targetID: Can be a Camera ID (triggers all outputs) or a specific Digital Output Entity ID.
// isCamera: Set to true if targetID is a Camera ID.
func (c *AvigilonClient) TriggerDigitalOutput(sessionID, targetID string, isCamera bool) error {
	payload := models.TriggerOutputPayload{
		Session:  sessionID,
		IsToggle: true, // Default to toggle (momentary trigger)
	}

	if isCamera {
		payload.ID = targetID
	} else {
		payload.EntityID = targetID
	}

	// Page 18: PUT /camera/commands/trigger-digital-output
	resp, err := c.HTTP.R().
		SetBody(payload).
		Put("/camera/commands/trigger-digital-output")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to trigger output: %s", resp.String())
	}

	return nil
}
