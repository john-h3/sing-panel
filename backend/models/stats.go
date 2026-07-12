package models

import "time"

// ServiceInfo represents sing-box service information
type ServiceInfo struct {
	StartTime time.Time `json:"startTime"`
}
