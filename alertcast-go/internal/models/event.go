package models

import "time"

type DeviceEvent struct {
	DeviceID    string    `json:"device_id"`
	Status      string    `json:"status"`       // online, offline, degraded
	Temperature float64   `json:"temperature"`  // Celsius
	Region      string    `json:"region"`
	Timestamp   time.Time `json:"ts"`
	Message     string    `json:"message,omitempty"`
}
