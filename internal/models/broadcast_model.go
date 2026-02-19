package models

import "time"

type BroadcastMessage struct {
	Alias       string    `json:"alias"`       // Device name
	Version     string    `json:"version"`     // Protocol version
	DeviceModel string    `json:"deviceModel"` // Device model
	DeviceType  string    `json:"deviceType"`  // Device type: mobile, desktop, web, headless, server
	Fingerprint string    `json:"fingerprint"` // Device fingerprint
	Port        int       `json:"port"`        // HTTP(S) server port
	Protocol    string    `json:"protocol"`    // Protocol: http or https
	Download    bool      `json:"download"`    // Whether download API is supported
	Announce    bool      `json:"announce"`    // Whether to announce presence
	LastSeen    time.Time `json:"-"`           // Last discovery time (local use only)
}
