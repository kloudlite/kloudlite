package analyzer

import (
	"os"
	"path/filepath"
	"strings"
)

// Language represents a detected programming language
type Language string

const (
	LangGo         Language = "go"
	LangJavaScript Language = "javascript"
	LangTypeScript Language = "typescript"
	LangPython     Language = "python"
	LangJava       Language = "java"
	LangRust       Language = "rust"
	LangC          Language = "c"
	LangCPP        Language = "cpp"
)

// extensionToLanguage maps file extensions to languages
var extensionToLanguage = map[string]Language{
	".go":   LangGo,
	".js":   LangJavaScript,
	".jsx":  LangJavaScript,
	".mjs":  LangJavaScript,
	".cjs":  LangJavaScript,
	".ts":   LangTypeScript,
	".tsx":  LangTypeScript,
	".mts":  LangTypeScript,
	".cts":  LangTypeScript,
	".py":   LangPython,
	".pyw":  LangPython,
	".pyx":  LangPython,
	".java": LangJava,
	".rs":   LangRust,
	".c":    LangC,
	".h":    LangC,
	".cpp":  LangCPP,
	".cc":   LangCPP,
	".cxx":  LangCPP,
	".hpp":  LangCPP,
	".hxx":  LangCPP,
}

// scanLanguageMapping maps scan language identifiers to actual languages
var scanLanguageMapping = map[string][]Language{
	"go":         {LangGo},
	"javascript": {LangJavaScript, LangTypeScript},
	"python":     {LangPython},
	"java":       {LangJava},
	"rust":       {LangRust},
	"c":          {LangC, LangCPP},
}

// DetectLanguages walks the workspace directory and detects programming languages
func DetectLanguages(workspaceDir string) ([]Language, error) {
	languageSet := make(map[Language]bool)

	err := filepath.WalkDir(workspaceDir, func(path string, d os.DirEntry, err error) error {
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

	if err != nil {
		return nil, err
	}

	// Convert set to slice
	languages := make([]Language, 0, len(languageSet))
	for lang := range languageSet {
		languages = append(languages, lang)
	}

	return languages, nil
}

// FilterScansForLanguages filters scan definitions to only include those applicable to detected languages
func FilterScansForLanguages(scans []ScanDefinition, detectedLanguages []Language) []ScanDefinition {
	if len(detectedLanguages) == 0 {
		// If no languages detected, only run generic scans (security + quality)
		return filterGenericScans(scans)
	}

	// Create lookup map for detected languages
	detectedSet := make(map[Language]bool)
	for _, lang := range detectedLanguages {
		detectedSet[lang] = true
	}

	filtered := make([]ScanDefinition, 0)

	for _, scan := range scans {
		if !scan.Enabled {
			continue
		}

		// If scan has no language restrictions, include it
		if len(scan.Languages) == 0 {
			filtered = append(filtered, scan)
			continue
		}

		// Check if any of the scan's target languages match detected languages
		if scanMatchesDetectedLanguages(scan.Languages, detectedSet) {
			filtered = append(filtered, scan)
		}
	}

	return filtered
}

// filterGenericScans returns only scans that apply to all languages
func filterGenericScans(scans []ScanDefinition) []ScanDefinition {
	filtered := make([]ScanDefinition, 0)
	for _, scan := range scans {
		if scan.Enabled && len(scan.Languages) == 0 {
			filtered = append(filtered, scan)
		}
	}
	return filtered
}

// scanMatchesDetectedLanguages checks if a scan should run based on detected languages
func scanMatchesDetectedLanguages(scanLanguages []string, detectedSet map[Language]bool) bool {
	for _, scanLang := range scanLanguages {
		// Map scan language identifier to actual languages
		if actualLanguages, ok := scanLanguageMapping[scanLang]; ok {
			for _, actualLang := range actualLanguages {
				if detectedSet[actualLang] {
					return true
				}
			}
		}
	}
	return false
}

// GetLanguageDisplayName returns a human-readable name for a language
func GetLanguageDisplayName(lang Language) string {
	names := map[Language]string{
		LangGo:         "Go",
		LangJavaScript: "JavaScript",
		LangTypeScript: "TypeScript",
		LangPython:     "Python",
		LangJava:       "Java",
		LangRust:       "Rust",
		LangC:          "C",
		LangCPP:        "C++",
	}
	if name, ok := names[lang]; ok {
		return name
	}
	return string(lang)
}

// LanguagesToStrings converts Language slice to string slice
func LanguagesToStrings(languages []Language) []string {
	result := make([]string, len(languages))
	for i, lang := range languages {
		result[i] = string(lang)
	}
	return result
}
