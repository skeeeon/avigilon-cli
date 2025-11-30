package client

import (
	"fmt"
	"avigilon-cli/pkg/models"
)

// GetSites fetches the list of ACC Sites (Clusters)
func (c *AvigilonClient) GetSites() ([]models.Site, error) {
	var respData models.SiteListResponse

	// Page 54: GET /sites
	resp, err := c.HTTP.R().
		SetResult(&respData).
		Get("/sites")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get sites: %s", resp.String())
	}

	return respData.Result.Sites, nil
}

// GetServers fetches the list of Servers in the current cluster
func (c *AvigilonClient) GetServers() ([]models.Server, error) {
	var respData models.ServerListResponse

	// Page 53: GET /server/ids
	resp, err := c.HTTP.R().
		SetResult(&respData).
		Get("/server/ids")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get servers: %s", resp.String())
	}

	return respData.Result.Servers, nil
}
