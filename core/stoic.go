package core

import (
	"encoding/json"
	"fmt"
)

// StoicPrompt returns the Daily Stoic heading for the given date.
// date must be in YYYY-MM-DD format. Returns empty string if not found.
func StoicPrompt(headingsJSON string, date string) string {
	if len(date) < 10 {
		return ""
	}
	// Extract MM-DD from YYYY-MM-DD
	key := date[5:7] + "-" + date[8:10]

	var headings map[string]string
	if err := json.Unmarshal([]byte(headingsJSON), &headings); err != nil {
		return ""
	}

	heading, ok := headings[key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s", heading)
}
