package deviceid

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	deviceIDFileName = "device-id"
)

// GetOrCreateDeviceID returns the persistent device UUID.
// If the device ID file doesn't exist, it generates a new UUID and saves it.
// The device ID is stored in ~/.kltun/device-id
func GetOrCreateDeviceID() (string, error) {
	deviceIDPath, err := getDeviceIDPath()
	if err != nil {
		return "", fmt.Errorf("failed to get device ID path: %w", err)
	}

	// Check if device ID file exists
	if _, err := os.Stat(deviceIDPath); err == nil {
		// File exists, read it
		data, err := os.ReadFile(deviceIDPath)
		if err != nil {
			return "", fmt.Errorf("failed to read device ID file: %w", err)
		}

		deviceID := strings.TrimSpace(string(data))

		// Validate it's a valid UUID
		if _, err := uuid.Parse(deviceID); err != nil {
			// Invalid UUID, regenerate
			return generateAndSaveDeviceID(deviceIDPath)
		}

		return deviceID, nil
	}

	// File doesn't exist, generate new UUID
	return generateAndSaveDeviceID(deviceIDPath)
}

// generateAndSaveDeviceID generates a new UUID and saves it to the specified path
func generateAndSaveDeviceID(deviceIDPath string) (string, error) {
	// Generate new UUID v4
	deviceID := uuid.New().String()

	// Create directory if it doesn't exist
	dir := filepath.Dir(deviceIDPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create device ID directory: %w", err)
	}

	// Write device ID to file with restrictive permissions
	if err := os.WriteFile(deviceIDPath, []byte(deviceID), 0600); err != nil {
		return "", fmt.Errorf("failed to write device ID file: %w", err)
	}

	return deviceID, nil
}

// getDeviceIDPath returns the path to the device ID file
func getDeviceIDPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".kltun", deviceIDFileName), nil
}

// GetDeviceIDPath returns the path where the device ID is stored (for informational purposes)
func GetDeviceIDPath() (string, error) {
	return getDeviceIDPath()
}
