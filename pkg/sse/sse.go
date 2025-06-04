package sse

import (
	"net/http"
)

type Server interface {
	// ServeHTTP streams events for a given channel (e.g., tenant or user)
	ServeHTTP(w http.ResponseWriter, r *http.Request, channel string)
	// Publish sends an event to the specified channel
	Publish(channel string, data []byte) error
}
