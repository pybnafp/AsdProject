package vo

// RagWorkerStatus represents the data returned for each worker in the status API
type RagWorkerStatus struct {
	ID                 int     `json:"id"`
	URL                string  `json:"url"`
	ActiveRequests     int     `json:"active_requests"`
	MaxCapacity        int     `json:"max_capacity"`
	IsAtMaxCapacity    bool    `json:"is_at_max_capacity"`
	CurrentLoadPercent float64 `json:"current_load_percent"`
}
