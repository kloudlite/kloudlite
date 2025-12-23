package analyzer

// ScanCategory represents the type of scan
type ScanCategory string

const (
	CategorySecurity ScanCategory = "security"
	CategoryQuality  ScanCategory = "quality"
	CategoryLanguage ScanCategory = "language"
)

// ScanDefinition defines a single scan type
type ScanDefinition struct {
	ID        string
	Name      string
	Category  ScanCategory
	CWE       []string // CWE references for security scans
	Languages []string // Empty = all languages, otherwise only run for these
	Prompt    string
	Enabled   bool
}

// ScanRegistry contains all available scans
var ScanRegistry = []ScanDefinition{
	// ============================================
	// SECURITY SCANS (OWASP/CWE-based)
	// ============================================

	{
		ID:       "SEC-01",
		Name:     "Secrets & Credentials",
		Category: CategorySecurity,
		CWE:      []string{"CWE-798", "CWE-259"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for hardcoded secrets and credentials.

Look for:
- API keys (AWS, GCP, Azure, Stripe, etc.)
- Passwords and database credentials
- Private keys and certificates
- JWT secrets and tokens
- Connection strings
- OAuth client secrets

Output ONLY valid JSON:
{"findings":[{"id":"SEC-01-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-02",
		Name:     "SQL Injection",
		Category: CategorySecurity,
		CWE:      []string{"CWE-89", "CWE-564"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for SQL injection vulnerabilities.

Look for:
- String concatenation in SQL queries
- Dynamic SQL without parameterization
- ORM raw query misuse
- Stored procedure injection

Output ONLY valid JSON:
{"findings":[{"id":"SEC-02-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-03",
		Name:     "Command Injection",
		Category: CategorySecurity,
		CWE:      []string{"CWE-78", "CWE-77"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for command injection vulnerabilities.

Look for:
- exec/system calls with user input
- Shell command execution
- Subprocess spawning with untrusted data
- eval() with external input

Output ONLY valid JSON:
{"findings":[{"id":"SEC-03-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-04",
		Name:     "XSS (Cross-Site Scripting)",
		Category: CategorySecurity,
		CWE:      []string{"CWE-79", "CWE-80"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for XSS vulnerabilities.

Look for:
- Reflected XSS (user input in response)
- Stored XSS (database content rendered)
- DOM-based XSS
- Template injection
- innerHTML with untrusted data

Output ONLY valid JSON:
{"findings":[{"id":"SEC-04-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-05",
		Name:     "Path Traversal",
		Category: CategorySecurity,
		CWE:      []string{"CWE-22", "CWE-23"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for path traversal vulnerabilities.

Look for:
- Directory traversal (../)
- File inclusion with user input
- Zip slip vulnerabilities
- Symlink attacks
- Unvalidated file paths

Output ONLY valid JSON:
{"findings":[{"id":"SEC-05-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-06",
		Name:     "SSRF",
		Category: CategorySecurity,
		CWE:      []string{"CWE-918"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for SSRF vulnerabilities.

Look for:
- URL fetching with user input
- Internal service access
- Cloud metadata endpoint access
- DNS rebinding risks

Output ONLY valid JSON:
{"findings":[{"id":"SEC-06-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-07",
		Name:     "Authentication Flaws",
		Category: CategorySecurity,
		CWE:      []string{"CWE-287", "CWE-306"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for authentication vulnerabilities.

Look for:
- Missing authentication checks
- Weak password requirements
- Session fixation
- JWT validation issues
- Credential stuffing risks

Output ONLY valid JSON:
{"findings":[{"id":"SEC-07-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-08",
		Name:     "Authorization Flaws",
		Category: CategorySecurity,
		CWE:      []string{"CWE-862", "CWE-863"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for authorization vulnerabilities.

Look for:
- Missing authorization checks
- IDOR (Insecure Direct Object Reference)
- Privilege escalation
- RBAC bypass
- Mass assignment

Output ONLY valid JSON:
{"findings":[{"id":"SEC-08-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-09",
		Name:     "Cryptography Issues",
		Category: CategorySecurity,
		CWE:      []string{"CWE-327", "CWE-328"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for cryptography vulnerabilities.

Look for:
- Weak algorithms (MD5, SHA1, DES)
- Insecure random number generation
- Hardcoded IVs/salts
- Missing encryption
- ECB mode usage

Output ONLY valid JSON:
{"findings":[{"id":"SEC-09-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-10",
		Name:     "Insecure Deserialization",
		Category: CategorySecurity,
		CWE:      []string{"CWE-502"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for insecure deserialization.

Look for:
- Pickle/unpickle with untrusted data
- YAML.load without safe_load
- JSON deserialization of untrusted data
- Object injection

Output ONLY valid JSON:
{"findings":[{"id":"SEC-10-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-11",
		Name:     "XML/XXE Vulnerabilities",
		Category: CategorySecurity,
		CWE:      []string{"CWE-611", "CWE-776"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for XML/XXE vulnerabilities.

Look for:
- XML External Entity injection
- Billion laughs attack
- XSLT injection
- Unsafe XML parsing

Output ONLY valid JSON:
{"findings":[{"id":"SEC-11-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-12",
		Name:     "LDAP Injection",
		Category: CategorySecurity,
		CWE:      []string{"CWE-90"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for LDAP injection vulnerabilities.

Look for:
- LDAP queries with user input
- Unescaped LDAP special characters
- Directory traversal in LDAP

Output ONLY valid JSON:
{"findings":[{"id":"SEC-12-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-13",
		Name:     "NoSQL Injection",
		Category: CategorySecurity,
		CWE:      []string{"CWE-943"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for NoSQL injection vulnerabilities.

Look for:
- MongoDB query injection
- Redis command injection
- Elasticsearch query injection
- Operator injection ($where, $regex)

Output ONLY valid JSON:
{"findings":[{"id":"SEC-13-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-14",
		Name:     "Regex DoS (ReDoS)",
		Category: CategorySecurity,
		CWE:      []string{"CWE-1333", "CWE-400"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for ReDoS vulnerabilities.

Look for:
- Catastrophic backtracking patterns
- Nested quantifiers
- Overlapping alternations
- User input in regex

Output ONLY valid JSON:
{"findings":[{"id":"SEC-14-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-15",
		Name:     "Race Conditions",
		Category: CategorySecurity,
		CWE:      []string{"CWE-362", "CWE-367"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for race condition vulnerabilities.

Look for:
- TOCTOU (time-of-check time-of-use)
- Double-checked locking issues
- Shared state without synchronization
- File system race conditions

Output ONLY valid JSON:
{"findings":[{"id":"SEC-15-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-16",
		Name:     "Sensitive Data Exposure",
		Category: CategorySecurity,
		CWE:      []string{"CWE-200", "CWE-532"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for sensitive data exposure.

Look for:
- Secrets in logs
- Verbose error messages
- Debug info in production
- Stack traces exposed
- PII in logs

Output ONLY valid JSON:
{"findings":[{"id":"SEC-16-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-17",
		Name:     "Insecure File Operations",
		Category: CategorySecurity,
		CWE:      []string{"CWE-73", "CWE-434"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for insecure file operation vulnerabilities.

Look for:
- Unrestricted file upload
- Insecure temp files
- Wrong file permissions
- Symlink vulnerabilities

Output ONLY valid JSON:
{"findings":[{"id":"SEC-17-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:        "SEC-18",
		Name:      "Memory Safety",
		Category:  CategorySecurity,
		CWE:       []string{"CWE-119", "CWE-120"},
		Languages: []string{"c", "cpp", "rust"},
		Enabled:   true,
		Prompt: `Analyze this codebase ONLY for memory safety vulnerabilities.

Look for:
- Buffer overflow
- Use-after-free
- Null pointer dereference
- Integer overflow
- Unsafe pointer arithmetic

Output ONLY valid JSON:
{"findings":[{"id":"SEC-18-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:        "SEC-19",
		Name:      "Prototype Pollution",
		Category:  CategorySecurity,
		CWE:       []string{"CWE-1321"},
		Languages: []string{"javascript", "typescript"},
		Enabled:   true,
		Prompt: `Analyze this codebase ONLY for prototype pollution vulnerabilities.

Look for:
- __proto__ manipulation
- constructor.prototype access
- Object.assign with untrusted data
- Deep merge vulnerabilities

Output ONLY valid JSON:
{"findings":[{"id":"SEC-19-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-20",
		Name:     "Open Redirect",
		Category: CategorySecurity,
		CWE:      []string{"CWE-601"},
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for open redirect vulnerabilities.

Look for:
- Unvalidated redirect URLs
- URL parameter manipulation
- Login redirect bypass

Output ONLY valid JSON:
{"findings":[{"id":"SEC-20-X","severity":"critical|high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	// ============================================
	// QUALITY SCANS (SonarQube/CodeClimate-based)
	// ============================================

	{
		ID:       "QUAL-01",
		Name:     "Cyclomatic Complexity",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for cyclomatic complexity issues.

Look for:
- Functions with complexity > 10
- Deeply nested conditionals
- Too many branches

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-01-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-02",
		Name:     "Cognitive Complexity",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for cognitive complexity issues.

Look for:
- Hard-to-understand code
- Mental overhead
- Complex control flow

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-02-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-03",
		Name:     "Code Duplication",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for code duplication.

Look for:
- Copy-paste code blocks
- Similar logic repeated
- Duplicate patterns

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-03-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-04",
		Name:     "Dead Code",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for dead code.

Look for:
- Unused variables
- Unreachable code
- Unused imports
- Commented-out code

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-04-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-05",
		Name:     "Error Handling",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for error handling issues.

Look for:
- Missing error checks
- Empty catch blocks
- Swallowed exceptions
- Panic risks

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-05-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-06",
		Name:     "Resource Leaks",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for resource leaks.

Look for:
- Unclosed files/connections
- Missing defer/finally
- Goroutine leaks
- Memory leaks

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-06-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-07",
		Name:     "Code Smells",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for code smells.

Look for:
- Long functions (>50 lines)
- Deep nesting (>4 levels)
- Too many parameters (>5)
- God classes/functions

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-07-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-08",
		Name:     "Naming Conventions",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for naming convention issues.

Look for:
- Single-letter variables (except i,j,k)
- Misleading names
- Inconsistent casing
- Abbreviations

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-08-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-09",
		Name:     "Magic Numbers/Strings",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for magic numbers and strings.

Look for:
- Hardcoded numeric values
- Unexplained literals
- Missing constants

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-09-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-10",
		Name:     "Documentation",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for documentation issues.

Look for:
- Missing function docs
- Outdated comments
- Complex undocumented logic
- Missing README sections

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-10-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-11",
		Name:     "Test Coverage Gaps",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for test coverage gaps.

Look for:
- Untested public functions
- Missing edge case tests
- No error path tests
- Integration test gaps

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-11-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-12",
		Name:     "Dependency Issues",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for dependency issues.

Look for:
- Deprecated packages
- Version conflicts
- Circular dependencies
- Heavy unused dependencies

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-12-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-13",
		Name:     "Performance Anti-patterns",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for performance anti-patterns.

Look for:
- N+1 queries
- Unnecessary allocations
- Blocking operations in async
- Inefficient algorithms

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-13-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-14",
		Name:     "Concurrency Issues",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for concurrency issues.

Look for:
- Data races
- Deadlock risks
- Missing synchronization
- Channel misuse

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-14-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-15",
		Name:     "API Design",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: `Analyze this codebase ONLY for API design issues.

Look for:
- Inconsistent APIs
- Breaking changes
- Missing validation
- Poor error responses

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-15-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N,"score":0-100}}`,
	},

	// ============================================
	// LANGUAGE-SPECIFIC SCANS
	// ============================================

	{
		ID:        "LANG-GO",
		Name:      "Go Best Practices",
		Category:  CategoryLanguage,
		Languages: []string{"go"},
		Enabled:   true,
		Prompt: `Analyze this Go codebase for language-specific issues.

Look for:
- Context handling issues
- Goroutine leaks
- Interface pollution
- Error wrapping issues
- Improper defer usage

Output ONLY valid JSON:
{"findings":[{"id":"LANG-GO-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:        "LANG-JS",
		Name:      "JavaScript/TypeScript Best Practices",
		Category:  CategoryLanguage,
		Languages: []string{"javascript", "typescript"},
		Enabled:   true,
		Prompt: `Analyze this JavaScript/TypeScript codebase for language-specific issues.

Look for:
- Async/await issues
- Type coercion bugs
- eval() usage
- Callback hell
- Promise anti-patterns

Output ONLY valid JSON:
{"findings":[{"id":"LANG-JS-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:        "LANG-PY",
		Name:      "Python Best Practices",
		Category:  CategoryLanguage,
		Languages: []string{"python"},
		Enabled:   true,
		Prompt: `Analyze this Python codebase for language-specific issues.

Look for:
- Type hint issues
- f-string injection
- Pickle usage
- Import issues
- Mutable default args

Output ONLY valid JSON:
{"findings":[{"id":"LANG-PY-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:        "LANG-JAVA",
		Name:      "Java Best Practices",
		Category:  CategoryLanguage,
		Languages: []string{"java"},
		Enabled:   true,
		Prompt: `Analyze this Java codebase for language-specific issues.

Look for:
- Null safety issues
- Stream misuse
- Reflection risks
- Serialization issues
- Resource management

Output ONLY valid JSON:
{"findings":[{"id":"LANG-JAVA-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:        "LANG-RUST",
		Name:      "Rust Best Practices",
		Category:  CategoryLanguage,
		Languages: []string{"rust"},
		Enabled:   true,
		Prompt: `Analyze this Rust codebase for language-specific issues.

Look for:
- Unsafe blocks
- unwrap() usage
- Lifetime issues
- Panic risks
- Clippy warnings

Output ONLY valid JSON:
{"findings":[{"id":"LANG-RUST-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},

	{
		ID:        "LANG-C",
		Name:      "C/C++ Best Practices",
		Category:  CategoryLanguage,
		Languages: []string{"c", "cpp"},
		Enabled:   true,
		Prompt: `Analyze this C/C++ codebase for language-specific issues.

Look for:
- Memory management issues
- Buffer handling
- Pointer arithmetic
- RAII violations
- Header issues

Output ONLY valid JSON:
{"findings":[{"id":"LANG-C-X","severity":"high|medium|low","file":"path","line":N,"title":"...","description":"...","recommendation":"..."}],"summary":{"count":N}}`,
	},
}

// GetEnabledScans returns all enabled scans
func GetEnabledScans() []ScanDefinition {
	var enabled []ScanDefinition
	for _, scan := range ScanRegistry {
		if scan.Enabled {
			enabled = append(enabled, scan)
		}
	}
	return enabled
}

// GetScansByCategory returns scans filtered by category
func GetScansByCategory(category ScanCategory) []ScanDefinition {
	var result []ScanDefinition
	for _, scan := range ScanRegistry {
		if scan.Category == category && scan.Enabled {
			result = append(result, scan)
		}
	}
	return result
}
