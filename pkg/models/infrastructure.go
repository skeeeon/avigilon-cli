package models

// --- Site Models ---

type SiteListResponse struct {
	Result struct {
		Sites []Site `json:"sites"`
	} `json:"result"`
}

type Site struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Add other fields if discovered during debug (e.g., "status", "version")
}

// --- Server Models ---

type ServerListResponse struct {
	Result struct {
		// Based on the endpoint /server/ids, this might be "servers" or just generic "ids"
		// We'll assume "servers" matches the Avigilon API style.
		Servers []Server `json:"servers"` 
	} `json:"result"`
}

type Server struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
