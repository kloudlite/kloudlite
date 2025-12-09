//go:build windows

package wintun

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

//go:embed dll/*.dll
var embeddedDLLs embed.FS

var (
	extractOnce sync.Once
	extractErr  error
)

// EnsureAvailable extracts wintun.dll to the executable's directory if not present.
// This must be called before any WireGuard TUN operations on Windows.
func EnsureAvailable() error {
	extractOnce.Do(func() {
		extractErr = extractWintunDLL()
	})
	return extractErr
}

func extractWintunDLL() error {
	// Determine which DLL to use based on architecture
	var dllName string
	switch runtime.GOARCH {
	case "amd64":
		dllName = "dll/wintun_amd64.dll"
	case "arm64":
		dllName = "dll/wintun_arm64.dll"
	default:
		return fmt.Errorf("unsupported architecture for wintun: %s", runtime.GOARCH)
	}

	// Read embedded DLL
	dllData, err := embeddedDLLs.ReadFile(dllName)
	if err != nil {
		return fmt.Errorf("failed to read embedded wintun.dll: %w", err)
	}

	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// Target path for wintun.dll (must be named wintun.dll for the library to find it)
	dllPath := filepath.Join(exeDir, "wintun.dll")

	// Check if DLL already exists and has correct size
	if info, err := os.Stat(dllPath); err == nil {
		if info.Size() == int64(len(dllData)) {
			// DLL exists and has correct size, assume it's correct
			return nil
		}
	}

	// Write the DLL
	if err := os.WriteFile(dllPath, dllData, 0644); err != nil {
		// If we can't write to exe directory (e.g., Program Files), try AppData
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			return fmt.Errorf("failed to write wintun.dll to %s: %w", dllPath, err)
		}

		kltunDir := filepath.Join(appData, "kltun")
		if err := os.MkdirAll(kltunDir, 0755); err != nil {
			return fmt.Errorf("failed to create kltun directory: %w", err)
		}

		dllPath = filepath.Join(kltunDir, "wintun.dll")
		if err := os.WriteFile(dllPath, dllData, 0644); err != nil {
			return fmt.Errorf("failed to write wintun.dll to %s: %w", dllPath, err)
		}

		// Add the directory to PATH so Windows can find the DLL
		currentPath := os.Getenv("PATH")
		os.Setenv("PATH", kltunDir+";"+currentPath)
	}

	return nil
}
