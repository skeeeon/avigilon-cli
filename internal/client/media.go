package client

import (
	"errors"
	"fmt"
)

// GetSnapshot downloads a JPEG snapshot for the given camera ID.
// Returns the binary byte slice of the image.
func (c *AvigilonClient) GetSnapshot(cameraID string) ([]byte, error) {
	// Page 44/45 parameters
	resp, err := c.HTTP.R().
		SetQueryParam("cameraId", cameraID).
		SetQueryParam("format", "jpeg"). // Request an image, not video
		Get("/media")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get snapshot: %s", resp.String())
	}

	// Basic validation to ensure we actually got an image
	contentType := resp.Header().Get("Content-Type")
	if len(resp.Body()) == 0 {
		return nil, errors.New("response body is empty")
	}
	
	// Avigilon usually returns "image/jpeg", but sometimes generic binary types
	// We'll trust the successful status code for now.
	fmt.Printf("Debug: Received Content-Type: %s, Size: %d bytes\n", contentType, len(resp.Body()))

	return resp.Body(), nil
}
