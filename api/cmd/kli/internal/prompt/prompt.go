package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReadLine reads a line from stdin with a prompt
func ReadLine(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// Confirm asks for y/n confirmation
func Confirm(message string) (bool, error) {
	for {
		input, err := ReadLine(fmt.Sprintf("%s [y/n]: ", message))
		if err != nil {
			return false, err
		}
		input = strings.ToLower(input)
		if input == "y" || input == "yes" {
			return true, nil
		}
		if input == "n" || input == "no" {
			return false, nil
		}
		fmt.Println("Please enter 'y' or 'n'")
	}
}

// ValidateSubdomain checks if a subdomain format is valid (client-side)
func ValidateSubdomain(subdomain string) (bool, string) {
	if len(subdomain) < 3 {
		return false, "Subdomain must be at least 3 characters"
	}
	if len(subdomain) > 63 {
		return false, "Subdomain must be at most 63 characters"
	}

	// Must start and end with alphanumeric
	if !isAlphanumeric(subdomain[0]) {
		return false, "Subdomain must start with a letter or number"
	}
	if !isAlphanumeric(subdomain[len(subdomain)-1]) {
		return false, "Subdomain must end with a letter or number"
	}

	// Check all characters are alphanumeric or hyphen
	for _, c := range subdomain {
		if !isAlphanumeric(byte(c)) && c != '-' {
			return false, "Subdomain can only contain letters, numbers, and hyphens"
		}
	}

	return true, ""
}

func isAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
