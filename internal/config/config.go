package config

import (
	"embed"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/meowrain/localsend-go/internal/utils/logger"
	"gopkg.in/yaml.v2"
)

//go:embed localsend.yaml
var embeddedConfig embed.FS

type Config struct {
	DeviceName   string `yaml:"device_name"`
	NameOfDevice string // Actual device name used in runtime
	SaveDir      string `yaml:"save_dir"`
	Functions    struct {
		HttpFileServer  bool `yaml:"http_file_server"`
		LocalSendServer bool `yaml:"local_send_server"`
	} `yaml:"functions"`
}

// random device name
var (
	adjectives = []string{
		"Happy", "Swift", "Silent", "Clever", "Brave",
		"Gentle", "Wise", "Calm", "Lucky", "Proud",
	}
	nouns = []string{
		"Phoenix", "Wolf", "Eagle", "Lion", "Owl",
		"Shark", "Tiger", "Bear", "Hawk", "Fox",
	}
)

var ConfigData Config

// random device name generator
func generateRandomName() string {
	localRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	adj := adjectives[localRand.Intn(len(adjectives))]
	noun := nouns[localRand.Intn(len(nouns))]
	return fmt.Sprintf("%s %s", adj, noun)
}

// LoadConfig loads configuration from the given path. If path is empty, it
// tries the default filesystem location. If reading from the filesystem fails,
// it falls back to the embedded config.
func LoadConfig(path string) {
	var bytes []byte
	var err error

	if path != "" {
		bytes, err = os.ReadFile(path)
		if err != nil {
			logger.Errorf("Failed to read config file %q: %v, falling back to embedded config", path, err)
			bytes, err = embeddedConfig.ReadFile("localsend.yaml")
			if err != nil {
				logger.Failedf("Can not read embedded config file: %v", err)
			}
		}
	} else {
		bytes, err = os.ReadFile("localsend.yaml")
		if err != nil {
			logger.Debug("Read localsend.yaml failed, using embedded config. Error: " + err.Error())
			bytes, err = embeddedConfig.ReadFile("localsend.yaml")
			if err != nil {
				logger.Failedf("Can not read embedded config file: %v", err)
			}
		}
	}

	if err := yaml.Unmarshal(bytes, &ConfigData); err != nil {
		logger.Failedf("Failed to parse config file: %v", err)
	}

	// Use configured device name if provided, otherwise generate a random one
	if ConfigData.DeviceName != "" {
		ConfigData.NameOfDevice = ConfigData.DeviceName
		logger.Debug("Using configured device name: " + ConfigData.NameOfDevice)
	} else {
		ConfigData.NameOfDevice = generateRandomName()
		logger.Debug("Using randomly generated device name: " + ConfigData.NameOfDevice)
	}

	// Default save directory
	if ConfigData.SaveDir == "" {
		ConfigData.SaveDir = "./uploads"
	}
}
