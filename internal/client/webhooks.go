package client

import (
	"fmt"
	"avigilon-cli/pkg/models"
)

// GetWebhooks lists all registered webhooks
func (c *AvigilonClient) GetWebhooks() ([]models.Webhook, error) {
	var respData models.WebhookListResponse

	resp, err := c.HTTP.R().
		SetResult(&respData).
		Get("/webhooks")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to list webhooks: %s", resp.String())
	}

	return respData.Result.Webhooks, nil
}

// CreateWebhook registers a new webhook with full configuration
// parameters including session ID (for body), auth token, and heartbeat settings.
func (c *AvigilonClient) CreateWebhook(sessionID, url, authToken string, topics []string, hbEnable bool, hbFreq int) error {
	payload := models.WebhookPayload{
		Session: sessionID, // Required field in the body
		Webhook: models.Webhook{
			URL:                 url,
			AuthenticationToken: authToken, // This must not be empty string
			Heartbeat: &models.Heartbeat{
				Enable:      hbEnable,
				FrequencyMs: hbFreq,
			},
			EventTopics: &models.EventTopics{
				Include: topics,
			},
		},
	}

	resp, err := c.HTTP.R().
		SetBody(payload).
		Post("/webhooks")

	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("failed to create webhook: %s", resp.String())
	}

	return nil
}

// DeleteWebhook removes a webhook by ID
func (c *AvigilonClient) DeleteWebhook(id string) error {
	// API requires passing IDs in the body for deletion
	payload := map[string][]string{
		"ids": {id},
	}

	resp, err := c.HTTP.R().
		SetBody(payload).
		Delete("/webhooks")

	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete webhook: %s", resp.String())
	}

	return nil
}
