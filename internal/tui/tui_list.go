package tui

import (
	"fmt"
	"time"

	"github.com/meowrain/localsend-go/internal/models"

	bubbletea "github.com/charmbracelet/bubbletea"
)

// SelectDevice displays a selectable device list using Bubble Tea and waits for user selection
func SelectDevice(updates <-chan []models.SendModel) (string, error) {
	// Create a buffered internal channel
	internalUpdates := make(chan []models.SendModel, 100)

	// Continuously read updates from the external channel in the background
	go func() {
		for devices := range updates {
			// Non-blocking send to internal channel
			select {
			case internalUpdates <- devices:
			default:
				// If the channel is full, drain and resend
				select {
				case <-internalUpdates:
				default:
				}
				internalUpdates <- devices
			}
		}
	}()

	// Create model and Bubble Tea program
	initModel := &model{
		devices:    []models.SendModel{},
		deviceMap:  make(map[string]models.SendModel),
		sortedKeys: make([]string, 0),
		cursor:     0,
		updates:    internalUpdates,
	}

	cmd := bubbletea.NewProgram(initModel)
	m, err := cmd.Run()
	if err != nil {
		return "", err
	}

	if m, ok := m.(model); ok && len(m.devices) > 0 {
		return m.devices[m.cursor].IP, nil
	}
	return "", nil
}

// model is the Bubble Tea model
type model struct {
	devices    []models.SendModel
	deviceMap  map[string]models.SendModel // Uses IP as key to store devices
	sortedKeys []string                    // Maintains a fixed display order
	cursor     int
	updates    <-chan []models.SendModel
}

// TickMsg triggers periodic updates
type TickMsg time.Time

// Init implements the Bubble Tea Init method
func (m model) Init() bubbletea.Cmd {
	return tick()
}

// tick fires once per second
func tick() bubbletea.Cmd {
	return bubbletea.Tick(time.Second, func(t time.Time) bubbletea.Msg {
		return TickMsg(t)
	})
}

// Update implements the Bubble Tea Update method
func (m model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case bubbletea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, bubbletea.Quit
		case "down", "j":
			if len(m.devices) > 0 {
				m.cursor = (m.cursor + 1) % len(m.devices) // Move down
			}
		case "up", "k":
			if len(m.devices) > 0 {
				m.cursor = (m.cursor - 1 + len(m.devices)) % len(m.devices) // Move up
			}
		case "enter":
			return m, bubbletea.Quit // Confirm selection
		}
	case TickMsg:
		select {
		case newDevices := <-m.updates:
			if m.deviceMap == nil {
				m.deviceMap = make(map[string]models.SendModel)
			}

			// Update device map
			changed := false
			for _, device := range newDevices {
				if _, exists := m.deviceMap[device.IP]; !exists {
					m.deviceMap[device.IP] = device
					m.sortedKeys = append(m.sortedKeys, device.IP)
					changed = true
				}
			}

			// Only update the device list when new devices are found
			if changed {
				m.devices = make([]models.SendModel, 0, len(m.deviceMap))
				for _, key := range m.sortedKeys {
					if device, ok := m.deviceMap[key]; ok {
						m.devices = append(m.devices, device)
					}
				}

				// Ensure cursor doesn't exceed device list bounds
				if m.cursor >= len(m.devices) {
					m.cursor = len(m.devices) - 1
				}
			}
		default:
		}
		return m, tick()
	}
	return m, nil
}

// View implements the Bubble Tea View method
func (m model) View() string {
	if len(m.devices) == 0 {
		return "Scanning Devices...\n\n Press Ctrl+C to exit"
	}

	s := "Found Devices:\n\n"
	for i, device := range m.devices {
		cursor := " " // No cursor by default
		if m.cursor == i {
			cursor = ">" // Selected cursor
		}
		s += fmt.Sprintf("%s %s (%s)\n", cursor, device.DeviceName, device.IP)
	}
	s += "\nUse arrow keys to navigate and enter to select. Press Ctrl+C to exit."
	return s
}
