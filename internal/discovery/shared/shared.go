package shared

import (
	"sync"

	"github.com/meowrain/localsend-go/internal/config"
	"github.com/meowrain/localsend-go/internal/models"
	"github.com/meowrain/localsend-go/internal/utils"
)

// Global device registry, mutex, and broadcast message

var (
	DiscoveredDevices = make(map[string]models.BroadcastMessage)
	DevicesMutex      sync.RWMutex
)

// https://github.com/localsend/protocol?tab=readme-ov-file#71-device-type
var Message models.BroadcastMessage

// InitMessage populates Message from the loaded config. Must be called after
// config.LoadConfig().
func InitMessage() {
	Message = models.BroadcastMessage{
		Alias:       config.ConfigData.NameOfDevice,
		Version:     "2.0",
		DeviceModel: utils.CheckOSType(),
		DeviceType:  "headless",      // CLI tool uses headless type
		Fingerprint: "random-string", // Should generate a unique fingerprint
		Port:        53317,
		Protocol:    "http",
		Download:    true,
		Announce:    true,
	}
}
