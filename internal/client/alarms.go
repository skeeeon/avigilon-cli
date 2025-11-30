package client

import (
	"fmt"
	"avigilon-cli/pkg/models"
)

// GetAlarms fetches active alarms using the plural /alarms endpoint
func (c *AvigilonClient) GetAlarms() ([]models.Alarm, error) {
	var respData models.AlarmListResponse

	// Page 3: GET /alarms (List/Search)
	resp, err := c.HTTP.R().
		SetResult(&respData).
		Get("/alarms")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get alarms: %s", resp.String())
	}

	return respData.Result.Alarms, nil
}

// UpdateAlarm performs an action on an alarm (ACKNOWLEDGE, PURGE, DISMISS)
// Uses the singular PUT /alarm endpoint
func (c *AvigilonClient) UpdateAlarm(sessionID, alarmID, action, note string) error {
	payload := models.AlarmUpdatePayload{
		Session: sessionID,
		ID:      alarmID,
		Action:  action,
		Note:    note,
	}

	// Page 1: PUT /alarm
	resp, err := c.HTTP.R().
		SetBody(payload).
		Put("/alarm")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to update alarm: %s", resp.String())
	}

	return nil
}
