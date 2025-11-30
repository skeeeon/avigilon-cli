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
