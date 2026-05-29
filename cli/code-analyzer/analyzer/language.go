package analyzer

import (
	"os"
	"path/filepath"
	"strings"
)

// extensionToLanguage maps file extensions to language strings
var extensionToLanguage = map[string]string{
	".go":    "go",
	".js":    "javascript",
	".jsx":   "javascript",
	".mjs":   "javascript",
	".cjs":   "javascript",
	".ts":    "typescript",
	".tsx":   "typescript",
	".mts":   "typescript",
	".cts":   "typescript",
	".py":    "python",
	".pyw":   "python",
	".pyx":   "python",
	".java":  "java",
	".rs":    "rust",
	".c":     "c",
	".h":     "c",
	".cpp":   "cpp",
	".cc":    "cpp",
	".cxx":   "cpp",
	".hpp":   "cpp",
	".hxx":   "cpp",
	".rb":    "ruby",
	".php":   "php",
	".kt":    "kotlin",
	".scala": "scala",
	".cs":    "csharp",
}

// ignoredDirs lists directories to skip during language detection
var ignoredDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	".git":         true,
	"__pycache__":  true,
	".nix-profile": true,
	"dist":         true,
	"build":        true,
	".next":        true,
	".venv":        true,
	"venv":         true,
}

// isIgnoredDir checks if a directory should be ignored
func isIgnoredDir(name string) bool {
	return ignoredDirs[name]
}

// DetectLanguages walks the workspace directory and detects programming languages
func DetectLanguages(workspaceDir string) []string {
	languageSet := make(map[string]bool)

	filepath.WalkDir(workspaceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}

		// Skip directories we don't want to analyze
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || isIgnoredDir(name) {
				return filepath.SkipDir
			}
			return nil
		}

		// Get file extension
		ext := strings.ToLower(filepath.Ext(path))
		if lang, ok := extensionToLanguage[ext]; ok {
			languageSet[lang] = true
		}

		return nil
	})

	// Convert set to slice
	languages := make([]string, 0, len(languageSet))
	for lang := range languageSet {
		languages = append(languages, lang)
	}

	return languages
}
