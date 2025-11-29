package wgkeys

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/curve25519"
)

const (
	privateKeyFileName = "wg-private-key"
	publicKeyFileName  = "wg-public-key"
)

// KeyPair holds a WireGuard key pair
type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

// GetOrCreateKeyPair returns the persistent WireGuard key pair.
// If the key files don't exist, it generates a new key pair and saves them.
// Keys are stored in ~/.kltun/wg-private-key and ~/.kltun/wg-public-key
func GetOrCreateKeyPair() (*KeyPair, error) {
	privateKeyPath, publicKeyPath, err := getKeyPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to get key paths: %w", err)
	}

	// Check if both key files exist
	privateKeyExists := fileExists(privateKeyPath)
	publicKeyExists := fileExists(publicKeyPath)

	if privateKeyExists && publicKeyExists {
		// Both files exist, read them
		privateKey, err := readKeyFile(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}

		publicKey, err := readKeyFile(publicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key: %w", err)
		}

		// Validate the keys
		if isValidKey(privateKey) && isValidKey(publicKey) {
			return &KeyPair{
				PrivateKey: privateKey,
				PublicKey:  publicKey,
			}, nil
		}

		// Keys are invalid, regenerate
		fmt.Println("Existing WireGuard keys are invalid, regenerating...")
	}

	// Generate new key pair
	return generateAndSaveKeyPair(privateKeyPath, publicKeyPath)
}

// generateAndSaveKeyPair generates a new WireGuard key pair and saves it
func generateAndSaveKeyPair(privateKeyPath, publicKeyPath string) (*KeyPair, error) {
	// Generate private key (32 random bytes)
	privateKeyBytes := make([]byte, 32)
	if _, err := randomBytes(privateKeyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Clamp the private key for Curve25519
	privateKeyBytes[0] &= 248
	privateKeyBytes[31] &= 127
	privateKeyBytes[31] |= 64

	// Derive public key from private key
	var publicKeyBytes [32]byte
	var privateKeyArray [32]byte
	copy(privateKeyArray[:], privateKeyBytes)
	curve25519.ScalarBaseMult(&publicKeyBytes, &privateKeyArray)

	// Encode keys to base64
	privateKey := base64.StdEncoding.EncodeToString(privateKeyBytes)
	publicKey := base64.StdEncoding.EncodeToString(publicKeyBytes[:])

	// Create directory if it doesn't exist
	dir := filepath.Dir(privateKeyPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create keys directory: %w", err)
	}

	// Write private key with restrictive permissions
	if err := os.WriteFile(privateKeyPath, []byte(privateKey), 0o600); err != nil {
		return nil, fmt.Errorf("failed to write private key: %w", err)
	}

	// Write public key
	if err := os.WriteFile(publicKeyPath, []byte(publicKey), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write public key: %w", err)
	}

	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// randomBytes fills the given byte slice with random data
func randomBytes(b []byte) (int, error) {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Read(b)
}

// getKeyPaths returns the paths to the private and public key files
func getKeyPaths() (privateKeyPath, publicKeyPath string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	kltunDir := filepath.Join(homeDir, ".kltun")
	return filepath.Join(kltunDir, privateKeyFileName), filepath.Join(kltunDir, publicKeyFileName), nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// readKeyFile reads a key file and returns the trimmed content
func readKeyFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// isValidKey checks if a key is valid (44 characters base64)
func isValidKey(key string) bool {
	if len(key) != 44 {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return false
	}
	return len(decoded) == 32
}

// GetPublicKey returns just the public key (useful for sending to server)
func GetPublicKey() (string, error) {
	keyPair, err := GetOrCreateKeyPair()
	if err != nil {
		return "", err
	}
	return keyPair.PublicKey, nil
}

// GetPrivateKey returns just the private key (for local WireGuard config)
func GetPrivateKey() (string, error) {
	keyPair, err := GetOrCreateKeyPair()
	if err != nil {
		return "", err
	}
	return keyPair.PrivateKey, nil
}
