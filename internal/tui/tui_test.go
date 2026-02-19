package tui

import (
	"testing"
	"time"

	"github.com/meowrain/localsend-go/internal/models"
)

// TestSelectDevice tests the SelectDevice function
func TestSelectDevice(t *testing.T) {
	// Create a device update channel
	updates := make(chan []models.SendModel)

	// Simulate device updates
	go func() {
		time.Sleep(1 * time.Second)
		updates <- []models.SendModel{
			{IP: "192.168.1.1", DeviceName: "Device 1"},
			{IP: "192.168.1.2", DeviceName: "Device 2"},
		}
		time.Sleep(1 * time.Second)
		updates <- []models.SendModel{
			{IP: "192.168.1.1", DeviceName: "Device 1"},
			{IP: "192.168.1.2", DeviceName: "Device 2"},
			{IP: "192.168.1.3", DeviceName: "Device 3"},
		}
	}()

	// Call the SelectDevice function
	ip, err := SelectDevice(updates)
	if err != nil {
		t.Fatalf("SelectDevice returned an error: %v", err)
	}

	// Check if the returned IP is in the simulated device list
	expectedIPs := map[string]bool{
		"192.168.1.1": true,
		"192.168.1.2": true,
		"192.168.1.3": true,
	}
	if !expectedIPs[ip] {
		t.Fatalf("SelectDevice returned an unexpected IP: %s", ip)
	}
}
