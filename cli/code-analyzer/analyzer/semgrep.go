package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SemgrepRunner executes Semgrep scans and parses results
type SemgrepRunner struct {
	config SemgrepConfig
	logger *zap.Logger
}

// SemgrepConfig holds Semgrep configuration
type SemgrepConfig struct {
	// Core security packs (always enabled)
	CorePacks []string

	// Language packs (auto-detected based on workspace)
	LanguagePacks []string

	// Framework packs (detected from package.json, requirements.txt, etc.)
	FrameworkPacks []string

	// Infrastructure packs (if Dockerfile, terraform, k8s files exist)
	InfraPacks []string

	// Custom rules path (optional)
	CustomRulesPath string

	// Excluded directories
	ExcludeDirs []string
}

// DefaultSemgrepConfig returns the default configuration with all standard packs
func DefaultSemgrepConfig() SemgrepConfig {
	return SemgrepConfig{
		CorePacks: []string{
			"p/security-audit",
			"p/owasp-top-ten",
			"p/cwe-top-25",
			"p/secrets",
			"p/ci",
		},
		ExcludeDirs: []string{
			"node_modules",
			"vendor",
			".git",
			"dist",
			"build",
			"__pycache__",
			".venv",
			"venv",
		},
	}
}

// NewSemgrepRunner creates a new Semgrep runner with default config
func NewSemgrepRunner(logger *zap.Logger) *SemgrepRunner {
	return &SemgrepRunner{
		config: DefaultSemgrepConfig(),
		logger: logger,
	}
}

// NewSemgrepRunnerWithConfig creates a new Semgrep runner with custom config
func NewSemgrepRunnerWithConfig(config SemgrepConfig, logger *zap.Logger) *SemgrepRunner {
	return &SemgrepRunner{
		config: config,
		logger: logger,
	}
}

// ScanResult represents the result of a Semgrep scan
type SemgrepScanResult struct {
	Findings []Finding
	Duration time.Duration
	Error    string
}

// Scan runs Semgrep on the given workspace directory
func (s *SemgrepRunner) Scan(ctx context.Context, workspaceDir string) (*SemgrepScanResult, error) {
	startTime := time.Now()

	// Detect languages and add appropriate rule packs
	languages := DetectLanguages(workspaceDir)
	s.addLanguagePacks(languages)

	// Detect frameworks and infrastructure
	s.detectFrameworks(workspaceDir)
	s.detectInfrastructure(workspaceDir)

	// Build semgrep command
	args := s.buildArgs(workspaceDir)

	s.logger.Info("Running Semgrep scan",
		zap.String("workspace", workspaceDir),
		zap.Strings("languages", languages),
		zap.Int("rule_packs", len(s.getAllPacks())),
	)

	// Create temp file for SARIF output
	tmpFile, err := os.CreateTemp("", "semgrep-*.sarif")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Add output flags
	args = append(args, "--sarif", "--sarif-output", tmpPath)

	// Run semgrep
	cmd := exec.CommandContext(ctx, "semgrep", args...)
	cmd.Dir = workspaceDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	duration := time.Since(startTime)

	// Semgrep returns exit code 1 if findings exist, which is not an error
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means findings were found (not an error)
			// Exit code > 1 means actual error
			if exitErr.ExitCode() > 1 {
				s.logger.Error("Semgrep execution failed",
					zap.Error(err),
					zap.String("stderr", stderr.String()),
				)
				return &SemgrepScanResult{
					Findings: []Finding{},
					Duration: duration,
					Error:    fmt.Sprintf("semgrep failed: %s", stderr.String()),
				}, nil
			}
		} else {
			return nil, fmt.Errorf("failed to run semgrep: %w", err)
		}
	}

	// Parse SARIF output
	sarifData, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SARIF output: %w", err)
	}

	findings, err := s.parseSARIF(sarifData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SARIF: %w", err)
	}

	s.logger.Info("Semgrep scan completed",
		zap.Duration("duration", duration),
		zap.Int("findings", len(findings)),
	)

	return &SemgrepScanResult{
		Findings: findings,
		Duration: duration,
	}, nil
}

// buildArgs constructs the semgrep command arguments
func (s *SemgrepRunner) buildArgs(workspaceDir string) []string {
	args := []string{"scan"}

	// Add all rule packs
	for _, pack := range s.getAllPacks() {
		args = append(args, "--config", pack)
	}

	// Add custom rules if specified
	if s.config.CustomRulesPath != "" {
		args = append(args, "--config", s.config.CustomRulesPath)
	}

	// Add exclusions
	for _, dir := range s.config.ExcludeDirs {
		args = append(args, "--exclude", dir)
	}

	// Add common exclusion patterns
	args = append(args,
		"--exclude", "*.min.js",
		"--exclude", "*.bundle.js",
		"--exclude", "*.map",
	)

	// Target directory
	args = append(args, workspaceDir)

	return args
}

// getAllPacks returns all configured rule packs
func (s *SemgrepRunner) getAllPacks() []string {
	var packs []string
	packs = append(packs, s.config.CorePacks...)
	packs = append(packs, s.config.LanguagePacks...)
	packs = append(packs, s.config.FrameworkPacks...)
	packs = append(packs, s.config.InfraPacks...)
	return packs
}

// addLanguagePacks adds language-specific rule packs based on detected languages
func (s *SemgrepRunner) addLanguagePacks(languages []string) {
	langPackMap := map[string][]string{
		"go":         {"p/golang"},
		"javascript": {"p/javascript", "p/nodejs"},
		"typescript": {"p/typescript", "p/javascript"},
		"python":     {"p/python"},
		"java":       {"p/java"},
		"rust":       {"p/rust"},
		"c":          {"p/c"},
		"cpp":        {"p/c"},
		"ruby":       {"p/ruby"},
		"php":        {"p/php"},
		"kotlin":     {"p/kotlin"},
		"scala":      {"p/scala"},
		"csharp":     {"p/csharp"},
	}

	packSet := make(map[string]bool)
	for _, lang := range languages {
		if packs, ok := langPackMap[lang]; ok {
			for _, pack := range packs {
				if !packSet[pack] {
					packSet[pack] = true
					s.config.LanguagePacks = append(s.config.LanguagePacks, pack)
				}
			}
		}
	}
}

// detectFrameworks detects frameworks and adds appropriate rule packs
func (s *SemgrepRunner) detectFrameworks(workspaceDir string) {
	// Check package.json for JS frameworks
	packageJSONPath := filepath.Join(workspaceDir, "package.json")
	if data, err := os.ReadFile(packageJSONPath); err == nil {
		content := string(data)
		if strings.Contains(content, "\"react\"") {
			s.config.FrameworkPacks = append(s.config.FrameworkPacks, "p/react")
		}
		if strings.Contains(content, "\"express\"") {
			s.config.FrameworkPacks = append(s.config.FrameworkPacks, "p/express")
		}
		if strings.Contains(content, "\"next\"") {
			s.config.FrameworkPacks = append(s.config.FrameworkPacks, "p/react")
		}
	}

	// Check requirements.txt for Python frameworks
	requirementsPath := filepath.Join(workspaceDir, "requirements.txt")
	if data, err := os.ReadFile(requirementsPath); err == nil {
		content := strings.ToLower(string(data))
		if strings.Contains(content, "django") {
			s.config.FrameworkPacks = append(s.config.FrameworkPacks, "p/django")
		}
		if strings.Contains(content, "flask") {
			s.config.FrameworkPacks = append(s.config.FrameworkPacks, "p/flask")
		}
	}

	// Check pom.xml for Java frameworks
	pomPath := filepath.Join(workspaceDir, "pom.xml")
	if data, err := os.ReadFile(pomPath); err == nil {
		content := string(data)
		if strings.Contains(content, "spring") {
			s.config.FrameworkPacks = append(s.config.FrameworkPacks, "p/spring")
		}
	}
}

// detectInfrastructure detects infrastructure files and adds appropriate rule packs
func (s *SemgrepRunner) detectInfrastructure(workspaceDir string) {
	// Check for Dockerfile
	if _, err := os.Stat(filepath.Join(workspaceDir, "Dockerfile")); err == nil {
		s.config.InfraPacks = append(s.config.InfraPacks, "p/dockerfile")
	}

	// Check for Terraform files
	matches, _ := filepath.Glob(filepath.Join(workspaceDir, "*.tf"))
	if len(matches) > 0 {
		s.config.InfraPacks = append(s.config.InfraPacks, "p/terraform")
	}

	// Check for Kubernetes files
	k8sPatterns := []string{
		filepath.Join(workspaceDir, "k8s", "*.yaml"),
		filepath.Join(workspaceDir, "k8s", "*.yml"),
		filepath.Join(workspaceDir, "kubernetes", "*.yaml"),
		filepath.Join(workspaceDir, "kubernetes", "*.yml"),
	}
	for _, pattern := range k8sPatterns {
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			s.config.InfraPacks = append(s.config.InfraPacks, "p/kubernetes")
			break
		}
	}
}

// SARIF types for parsing Semgrep output

// SARIFReport represents the top-level SARIF structure
type SARIFReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SARIFRun `json:"runs"`
}

// SARIFRun represents a single run in SARIF
type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

// SARIFTool represents the tool information
type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

// SARIFDriver represents the driver/tool information
type SARIFDriver struct {
	Name            string      `json:"name"`
	SemanticVersion string      `json:"semanticVersion"`
	Rules           []SARIFRule `json:"rules"`
}

// SARIFRule represents a rule definition
type SARIFRule struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	ShortDescription SARIFMessage           `json:"shortDescription"`
	FullDescription  SARIFMessage           `json:"fullDescription"`
	DefaultConfig    map[string]interface{} `json:"defaultConfiguration"`
	Properties       map[string]interface{} `json:"properties"`
}

// SARIFResult represents a single finding
type SARIFResult struct {
	RuleID     string                 `json:"ruleId"`
	Level      string                 `json:"level"` // error, warning, note
	Message    SARIFMessage           `json:"message"`
	Locations  []SARIFLocation        `json:"locations"`
	Fixes      []SARIFFix             `json:"fixes,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// SARIFMessage represents a message
type SARIFMessage struct {
	Text string `json:"text"`
}

// SARIFLocation represents a code location
type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

// SARIFPhysicalLocation represents a physical location in code
type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region"`
}

// SARIFArtifactLocation represents the file location
type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

// SARIFRegion represents a region in a file
type SARIFRegion struct {
	StartLine   int           `json:"startLine"`
	StartColumn int           `json:"startColumn,omitempty"`
	EndLine     int           `json:"endLine,omitempty"`
	EndColumn   int           `json:"endColumn,omitempty"`
	Snippet     *SARIFSnippet `json:"snippet,omitempty"`
}

// SARIFSnippet represents a code snippet
type SARIFSnippet struct {
	Text string `json:"text"`
}

// SARIFFix represents a suggested fix
type SARIFFix struct {
	Description SARIFMessage `json:"description"`
}

// parseSARIF parses SARIF output and converts to Findings
func (s *SemgrepRunner) parseSARIF(data []byte) ([]Finding, error) {
	var report SARIFReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SARIF: %w", err)
	}

	if len(report.Runs) == 0 {
		return []Finding{}, nil
	}

	// Build rule map for additional metadata
	ruleMap := make(map[string]SARIFRule)
	for _, rule := range report.Runs[0].Tool.Driver.Rules {
		ruleMap[rule.ID] = rule
	}

	var findings []Finding
	for _, result := range report.Runs[0].Results {
		finding := s.resultToFinding(result, ruleMap)
		findings = append(findings, finding)
	}

	return findings, nil
}

// resultToFinding converts a SARIF result to a Finding
func (s *SemgrepRunner) resultToFinding(result SARIFResult, ruleMap map[string]SARIFRule) Finding {
	finding := Finding{
		ID:          result.RuleID,
		Title:       result.Message.Text,
		Description: result.Message.Text,
		Severity:    s.mapSeverity(result.Level),
	}

	// Extract location
	if len(result.Locations) > 0 {
		loc := result.Locations[0].PhysicalLocation
		finding.File = loc.ArtifactLocation.URI
		finding.Line = loc.Region.StartLine

		// Extract code snippet as evidence
		if loc.Region.Snippet != nil {
			finding.Evidence = loc.Region.Snippet.Text
		}
	}

	// Extract category from rule ID (e.g., "go.lang.security.sql-injection" -> "security")
	finding.Category = s.extractCategory(result.RuleID)

	// Extract metadata from rule properties
	if rule, ok := ruleMap[result.RuleID]; ok {
		if cwe, ok := rule.Properties["cwe"].(string); ok {
			finding.CWE = cwe
		} else if cweList, ok := rule.Properties["cwe"].([]interface{}); ok && len(cweList) > 0 {
			if cweStr, ok := cweList[0].(string); ok {
				finding.CWE = cweStr
			}
		}

		if owasp, ok := rule.Properties["owasp"].(string); ok {
			finding.OWASP = owasp
		} else if owaspList, ok := rule.Properties["owasp"].([]interface{}); ok && len(owaspList) > 0 {
			if owaspStr, ok := owaspList[0].(string); ok {
				finding.OWASP = owaspStr
			}
		}

		// Use full description if available
		if rule.FullDescription.Text != "" {
			finding.Description = rule.FullDescription.Text
		}
	}

	// Extract recommendation from fixes if available
	if len(result.Fixes) > 0 {
		finding.Recommendation = result.Fixes[0].Description.Text
	}

	// Set tags based on rule ID and severity
	finding.Tags = s.generateTags(result, finding)

	return finding
}

// mapSeverity maps SARIF level to our severity
func (s *SemgrepRunner) mapSeverity(level string) string {
	switch level {
	case "error":
		return "high"
	case "warning":
		return "medium"
	case "note":
		return "low"
	default:
		return "medium"
	}
}

// extractCategory extracts category from Semgrep rule ID
func (s *SemgrepRunner) extractCategory(ruleID string) string {
	parts := strings.Split(ruleID, ".")
	if len(parts) >= 3 {
		// e.g., "go.lang.security.sql-injection" -> "security"
		category := parts[2]
		if category == "security" || category == "correctness" || category == "performance" {
			return category
		}
	}
	// Default to security for security-related rules
	if strings.Contains(strings.ToLower(ruleID), "security") ||
		strings.Contains(strings.ToLower(ruleID), "injection") ||
		strings.Contains(strings.ToLower(ruleID), "xss") ||
		strings.Contains(strings.ToLower(ruleID), "secret") {
		return "security"
	}
	return "quality"
}

// generateTags generates tags for a finding
func (s *SemgrepRunner) generateTags(result SARIFResult, finding Finding) []string {
	var tags []string

	// Add confidence tag - Semgrep findings are deterministic
	tags = append(tags, "CONFIRMED")

	// Add evidence tag
	tags = append(tags, "HAS_CODE_PATTERN")

	// Add category tag based on rule ID
	ruleID := strings.ToLower(result.RuleID)
	switch {
	case strings.Contains(ruleID, "injection") || strings.Contains(ruleID, "sql"):
		tags = append(tags, "SEC_INJECTION")
	case strings.Contains(ruleID, "xss"):
		tags = append(tags, "SEC_XSS")
	case strings.Contains(ruleID, "secret") || strings.Contains(ruleID, "credential") || strings.Contains(ruleID, "password"):
		tags = append(tags, "SEC_CRYPTO")
	case strings.Contains(ruleID, "ssrf"):
		tags = append(tags, "SEC_SSRF")
	case strings.Contains(ruleID, "auth"):
		tags = append(tags, "SEC_AUTH")
	case strings.Contains(ruleID, "command"):
		tags = append(tags, "SEC_INJECTION")
	case strings.Contains(ruleID, "path") || strings.Contains(ruleID, "traversal"):
		tags = append(tags, "SEC_ACCESS_CONTROL")
	case strings.Contains(ruleID, "xxe") || strings.Contains(ruleID, "xml"):
		tags = append(tags, "SEC_INJECTION")
	case strings.Contains(ruleID, "deserial"):
		tags = append(tags, "SEC_DESERIALIZATION")
	case finding.Category == "security":
		tags = append(tags, "SEC_MISCONFIG")
	default:
		tags = append(tags, "QUAL_MAINTAINABILITY")
	}

	return tags
}
