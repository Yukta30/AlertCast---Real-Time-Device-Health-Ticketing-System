package rules

import (
	"fmt"
	"strings"

	"alertcast/internal/models"
)

// Evaluate inspects a DeviceEvent and decides whether to open a ticket.
// Returns (severity, reason, shouldOpenTicket).
func Evaluate(e models.DeviceEvent) (string, string, bool) {
	status := strings.ToLower(e.Status)

	// Rule 1: Offline is critical.
	if status == "offline" {
		return "critical", "Device offline", true
	}

	// Rule 2: Temperature thresholds.
	if e.Temperature >= 85 {
		return "high", fmt.Sprintf("High temperature %.1f°C", e.Temperature), true
	}
	if e.Temperature >= 75 {
		return "medium", fmt.Sprintf("Elevated temperature %.1f°C", e.Temperature), true
	}

	// Rule 3: Degraded status (medium).
	if status == "degraded" {
		return "medium", "Device degraded", true
	}

	// Otherwise, no ticket
	return "", "", false
}
