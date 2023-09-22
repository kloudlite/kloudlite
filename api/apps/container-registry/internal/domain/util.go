package domain

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

func intnonce(size int) string {
	from := "0123456789"
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = from[rand.Intn(len(from))]
	}
	return string(buf)
}

func generateUserNames(input string, size int) []string {
	usernames := []string{}

	for i := 0; i < size; i++ {
		usernames = append(usernames, generateUsername(input))
	}

	return usernames
}

func generateUsername(input string) string {
	// Convert input to lowercase
	lowerInput := strings.ToLower(input)

	// Remove characters that are not a-z, 0-9, space or _
	cleaned := regexp.MustCompile("[^a-z0-9_ ]").ReplaceAllString(lowerInput, "")

	// Replace spaces with underscores
	cleaned = strings.ReplaceAll(cleaned, " ", "_")

	// Ensure the string is long enough. If not, pad with underscores
	for len(cleaned) < 5 {
		cleaned += "_"
	}

	// Ensure the string is not too long
	if len(cleaned) > 12 {
		cleaned = cleaned[:12]
	}

	// Verify that the string matches the required pattern
	match, _ := regexp.MatchString("^([a-z])([a-z0-9_])+$", cleaned)
	if !match {
		return generateUsername("user")
	}

	return cleaned + intnonce(4)
}

func isValidUserName(name string) error {
	pattern := `^([a-z])([a-z0-9_])+$`

	re := regexp.MustCompile(pattern)

	if !re.MatchString(name) {
		return fmt.Errorf("invalid username, must be lowercase alphanumeric with underscore")
	}

	return nil
}
