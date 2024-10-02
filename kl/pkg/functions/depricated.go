package functions

import (
	"fmt"
	"io"
	"os"
)

func MultiLine(s string, length int) string {
	resp := ""

	needToBreak := false
	for i, k := range s {
		resp += string(k)

		if (i+1)%length == 0 {
			needToBreak = true
		}

		if needToBreak && string(k) == " " {
			resp += "\n"
			needToBreak = false
		}

	}

	return resp
}

func CopyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return Errorf("Failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return Errorf("Failed to create destination file: %v", err)
	}
	defer destFile.Close()

	// Copy the contents from source file to destination file
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return Errorf("Failed to copy file contents: %v", err)
	}

	// Flush file contents to disk
	err = destFile.Sync()
	if err != nil {
		return Errorf("Failed to flush file contents: %v", err)
	}

	return nil
}

func RemoveFromArray(target string, arr []string) []string {
	var result []string
	for _, s := range arr {
		if s != target {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func Truncate(str string, length int) string {
	if len(str) < length {
		return str
	}

	return fmt.Sprintf("%s...", str[0:length])
}
