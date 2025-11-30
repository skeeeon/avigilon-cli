package client

import (
	"fmt"
	"time"

	"avigilon-cli/pkg/models"
)

// Avigilon strict ISO 8601 format with milliseconds and Z suffix
const AvigilonTimeFormat = "2006-01-02T15:04:05.000Z"

// GetEvents searches for events on a specific server within a time range.
func (c *AvigilonClient) GetEvents(serverID string, from time.Time, to time.Time, topics []string) ([]models.Event, error) {
	var respData models.EventListResponse

	req := c.HTTP.R().
		SetQueryParam("queryType", "TIME_RANGE").
		SetQueryParam("serverId", serverID).
		SetQueryParam("from", from.UTC().Format(AvigilonTimeFormat))

	if !to.IsZero() {
		req.SetQueryParam("to", to.UTC().Format(AvigilonTimeFormat))
	}

	// FIX: Use req.QueryParam.Add() to append multiple values for the same key.
	// SetQueryParam() overwrites, which is why only the last topic was working previously.
	for _, t := range topics {
		if t != "" {
			req.QueryParam.Add("eventTopics", t)
		}
	}

	// Default limit
	req.SetQueryParam("limit", "100")

	resp, err := req.
		SetResult(&respData).
		Get("/events/search")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to search events on server %s: %s", serverID, resp.String())
	}

	return respData.Result.Events, nil
}
