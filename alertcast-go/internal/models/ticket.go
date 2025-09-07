package models

import "time"

type Ticket struct {
	ID        int64     `json:"id"`
	DeviceID  string    `json:"device_id"`
	Severity  string    `json:"severity"` // critical, high, medium, low
	Reason    string    `json:"reason"`
	EventJSON string    `json:"event_json"`
	CreatedAt time.Time `json:"created_at"`
}
