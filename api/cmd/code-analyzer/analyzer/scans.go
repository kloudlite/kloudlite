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

// StandardPromptRules is prepended to all scan prompts
const StandardPromptRules = `CRITICAL ANALYSIS RULES:
1. Report ONLY confirmed issues with concrete evidence
2. DO NOT report theoretical, potential, or speculative issues
3. DO NOT report suggestions, improvements, or best practices
4. If no confirmed issues exist, return {"findings":[],"summary":{"count":0}}
5. Each finding MUST have exact file:line location and evidence

`

// ScanRegistry contains all available scans
var ScanRegistry = []ScanDefinition{
	// ============================================
	// SECURITY SCANS (OWASP Top 10 + CWE Top 25)
	// ============================================

	{
		ID:       "SEC-01",
		Name:     "Secrets & Credentials",
		Category: CategorySecurity,
		CWE:      []string{"CWE-798", "CWE-259"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Hardcoded Secrets (OWASP A05, CWE-798)

REPORT ONLY if you find these CONFIRMED patterns:
- Actual API keys with valid format (e.g., AKIA*, sk-*, ghp_*)
- Passwords assigned to variables (password = "...")
- Private keys (-----BEGIN RSA PRIVATE KEY-----)
- Database connection strings with embedded credentials
- JWT secrets assigned in code

DO NOT REPORT:
- Environment variable references (os.Getenv, process.env)
- Configuration file placeholders (<YOUR_KEY_HERE>)
- Test/mock credentials clearly labeled as such
- Public keys (only private keys are issues)
- Empty strings or placeholder values

Output ONLY valid JSON:
{"findings":[{"id":"SEC-01-X","severity":"critical|high","file":"path","line":N,"title":"Hardcoded [type]","description":"Found [credential type] at line N: [masked evidence]","recommendation":"Move to environment variable or secrets manager"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-02",
		Name:     "SQL Injection",
		Category: CategorySecurity,
		CWE:      []string{"CWE-89", "CWE-564"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: SQL Injection (OWASP A03, CWE-89)

REPORT ONLY if you find these CONFIRMED patterns:
- String concatenation building SQL: "SELECT * FROM users WHERE id=" + userInput
- fmt.Sprintf/printf in SQL: fmt.Sprintf("SELECT * FROM users WHERE id=%s", id)
- Raw SQL with unsanitized variables directly interpolated
- Dynamic table/column names from user input

DO NOT REPORT:
- Parameterized queries: db.Query("SELECT * FROM users WHERE id=?", id)
- Prepared statements with placeholders ($1, :param, ?)
- ORM methods with proper parameter binding
- Static SQL strings without user input
- String concatenation with only constants

Output ONLY valid JSON:
{"findings":[{"id":"SEC-02-X","severity":"critical|high","file":"path","line":N,"title":"SQL Injection","description":"User input [variable] concatenated into SQL query without parameterization","recommendation":"Use parameterized queries"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-03",
		Name:     "Command Injection",
		Category: CategorySecurity,
		CWE:      []string{"CWE-78", "CWE-77"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Command Injection (OWASP A03, CWE-78)

REPORT ONLY if you find these CONFIRMED patterns:
- exec.Command with user input in command or arguments
- os.system/subprocess with unsanitized user input
- Shell execution (sh -c) with user-controlled strings
- eval() with user input (in JS/Python)

DO NOT REPORT:
- Hardcoded commands without user input
- Commands with validated/whitelisted arguments only
- exec.Command with constant strings
- System calls for internal operations without external input

Output ONLY valid JSON:
{"findings":[{"id":"SEC-03-X","severity":"critical","file":"path","line":N,"title":"Command Injection","description":"User input [variable] passed to [function] without sanitization","recommendation":"Validate input against whitelist or avoid shell execution"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-04",
		Name:     "XSS (Cross-Site Scripting)",
		Category: CategorySecurity,
		CWE:      []string{"CWE-79", "CWE-80"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Cross-Site Scripting (OWASP A03, CWE-79)

REPORT ONLY if you find these CONFIRMED patterns:
- innerHTML/outerHTML with user input
- document.write() with unsanitized data
- Unescaped template rendering: {{.UserInput}} without escaping
- Response.Write with unsanitized user input in HTML context
- dangerouslySetInnerHTML with user data

DO NOT REPORT:
- Properly escaped template output (html/template in Go)
- React JSX expressions (auto-escaped)
- User input in non-HTML contexts (JSON responses)
- Static HTML without user input

Output ONLY valid JSON:
{"findings":[{"id":"SEC-04-X","severity":"high","file":"path","line":N,"title":"XSS Vulnerability","description":"User input rendered without escaping via [method]","recommendation":"Escape output or use safe templating"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-05",
		Name:     "Path Traversal",
		Category: CategorySecurity,
		CWE:      []string{"CWE-22", "CWE-23"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Path Traversal (OWASP A01, CWE-22)

REPORT ONLY if you find these CONFIRMED patterns:
- File operations with user input: os.Open(userInput)
- Path.Join with user input without validation: filepath.Join(base, userInput)
- No check for ".." in user-provided paths
- Serving files based on user input without path validation

DO NOT REPORT:
- Path operations with validated/cleaned paths
- filepath.Clean() used before file operations
- Paths checked against base directory
- Static file paths without user input

Output ONLY valid JSON:
{"findings":[{"id":"SEC-05-X","severity":"high","file":"path","line":N,"title":"Path Traversal","description":"User input [variable] used in file path without validation","recommendation":"Validate path is within allowed directory"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-06",
		Name:     "SSRF",
		Category: CategorySecurity,
		CWE:      []string{"CWE-918"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Server-Side Request Forgery (OWASP A10, CWE-918)

REPORT ONLY if you find these CONFIRMED patterns:
- HTTP client requests with user-controlled URLs
- URL fetching without domain validation
- Redirect following to user-supplied URLs
- Internal service access via user input

DO NOT REPORT:
- Requests to hardcoded/whitelisted URLs only
- URLs validated against allowlist
- Internal API calls without user input
- Webhook URLs stored in database (runtime configured)

Output ONLY valid JSON:
{"findings":[{"id":"SEC-06-X","severity":"high","file":"path","line":N,"title":"SSRF Vulnerability","description":"HTTP request to user-controlled URL [variable] without validation","recommendation":"Validate URL against allowlist"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-07",
		Name:     "Authentication Bypass",
		Category: CategorySecurity,
		CWE:      []string{"CWE-287", "CWE-306"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Authentication Issues (OWASP A07, CWE-287)

REPORT ONLY if you find these CONFIRMED patterns:
- Endpoints without authentication middleware when they should have it
- JWT/session validation that can be bypassed
- Password comparison using == instead of constant-time compare
- Missing authentication check on sensitive operations

DO NOT REPORT:
- Public endpoints intentionally unauthenticated
- Authentication handled by middleware/framework
- Proper session validation in place
- Design decisions about what requires auth (you don't know the requirements)

Output ONLY valid JSON:
{"findings":[{"id":"SEC-07-X","severity":"critical|high","file":"path","line":N,"title":"Authentication Issue","description":"[Specific issue found with evidence]","recommendation":"[Specific fix]"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-08",
		Name:     "Authorization Bypass",
		Category: CategorySecurity,
		CWE:      []string{"CWE-862", "CWE-863"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Authorization Issues (OWASP A01, CWE-862)

REPORT ONLY if you find these CONFIRMED patterns:
- Direct object access without ownership check
- User ID from JWT/session not verified against resource owner
- Role check missing on admin operations
- Mass assignment allowing role/privilege fields

DO NOT REPORT:
- Authorization handled by middleware
- Ownership checks present in code
- Role-based access properly implemented
- Design decisions about access control

Output ONLY valid JSON:
{"findings":[{"id":"SEC-08-X","severity":"high","file":"path","line":N,"title":"Authorization Issue","description":"[Specific issue with evidence]","recommendation":"[Specific fix]"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-09",
		Name:     "Weak Cryptography",
		Category: CategorySecurity,
		CWE:      []string{"CWE-327", "CWE-328"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Cryptographic Issues (OWASP A02, CWE-327)

REPORT ONLY if you find these CONFIRMED patterns:
- MD5/SHA1 used for password hashing (not HMAC)
- DES/3DES/RC4 encryption
- ECB mode block cipher
- Hardcoded encryption keys/IVs
- Math.random()/rand() for security purposes

DO NOT REPORT:
- MD5/SHA1 for checksums or non-security purposes
- Proper algorithms (bcrypt, argon2, AES-GCM)
- Keys loaded from environment/config
- Secure random generators (crypto/rand)

Output ONLY valid JSON:
{"findings":[{"id":"SEC-09-X","severity":"high","file":"path","line":N,"title":"Weak Cryptography","description":"[Algorithm] used for [purpose] at line N","recommendation":"Use [recommended algorithm]"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-10",
		Name:     "Insecure Deserialization",
		Category: CategorySecurity,
		CWE:      []string{"CWE-502"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Insecure Deserialization (OWASP A08, CWE-502)

REPORT ONLY if you find these CONFIRMED patterns:
- pickle.loads() with untrusted data
- yaml.load() without safe_load
- Java ObjectInputStream with untrusted data
- PHP unserialize() with user input
- eval() on serialized data

DO NOT REPORT:
- JSON parsing (generally safe)
- yaml.safe_load()
- Deserialization of trusted internal data
- Type-safe deserialization with schemas

Output ONLY valid JSON:
{"findings":[{"id":"SEC-10-X","severity":"critical","file":"path","line":N,"title":"Insecure Deserialization","description":"[Function] deserializes untrusted data from [source]","recommendation":"Use safe deserialization or validate input"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-11",
		Name:     "XXE Injection",
		Category: CategorySecurity,
		CWE:      []string{"CWE-611", "CWE-776"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: XML External Entity (OWASP A05, CWE-611)

REPORT ONLY if you find these CONFIRMED patterns:
- XML parsing without disabling external entities
- DTD processing enabled with user input
- XSLT processing with user-controlled stylesheets

DO NOT REPORT:
- XML parsing with external entities disabled
- JSON parsing (not affected)
- XML generation without parsing user XML

Output ONLY valid JSON:
{"findings":[{"id":"SEC-11-X","severity":"high","file":"path","line":N,"title":"XXE Vulnerability","description":"XML parser at line N allows external entities","recommendation":"Disable external entity processing"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-12",
		Name:     "LDAP Injection",
		Category: CategorySecurity,
		CWE:      []string{"CWE-90"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: LDAP Injection (CWE-90)

REPORT ONLY if you find these CONFIRMED patterns:
- LDAP filter string concatenation with user input
- User input in LDAP DN without escaping
- ldap.search() with unsanitized filter

DO NOT REPORT:
- No LDAP library usage in codebase
- Properly escaped LDAP queries
- LDAP operations with internal data only

Output ONLY valid JSON:
{"findings":[{"id":"SEC-12-X","severity":"high","file":"path","line":N,"title":"LDAP Injection","description":"User input in LDAP query without escaping","recommendation":"Escape LDAP special characters"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-13",
		Name:     "NoSQL Injection",
		Category: CategorySecurity,
		CWE:      []string{"CWE-943"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: NoSQL Injection (CWE-943)

REPORT ONLY if you find these CONFIRMED patterns:
- MongoDB query with user input allowing operator injection
- $where clause with user input
- User input in Redis commands
- Dynamic query building without type checking

DO NOT REPORT:
- Queries with type-validated input
- ORM/ODM with proper parameter binding
- Static queries without user input

Output ONLY valid JSON:
{"findings":[{"id":"SEC-13-X","severity":"high","file":"path","line":N,"title":"NoSQL Injection","description":"User input allows [operator/query] injection","recommendation":"Validate input types and sanitize"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-14",
		Name:     "ReDoS",
		Category: CategorySecurity,
		CWE:      []string{"CWE-1333", "CWE-400"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Regular Expression DoS (CWE-1333)

REPORT ONLY if you find these CONFIRMED patterns:
- Regex with nested quantifiers: (a+)+, (a|a)+
- Overlapping alternations with repetition
- User input used directly as regex pattern
- Known vulnerable patterns: .*.*

DO NOT REPORT:
- Simple regexes without nested quantifiers
- Regex with timeout protection
- Compiled static patterns

Output ONLY valid JSON:
{"findings":[{"id":"SEC-14-X","severity":"medium","file":"path","line":N,"title":"ReDoS Vulnerability","description":"Regex pattern [pattern] vulnerable to catastrophic backtracking","recommendation":"Simplify pattern or add timeout"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-15",
		Name:     "Race Conditions",
		Category: CategorySecurity,
		CWE:      []string{"CWE-362", "CWE-367"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Race Conditions (CWE-362)

REPORT ONLY if you find these CONFIRMED patterns:
- TOCTOU: check then use without locking (file exists check then open)
- Shared variable modification without synchronization
- Double-checked locking without volatile/atomic
- File operations without proper locking

DO NOT REPORT:
- Properly synchronized code (mutex, atomic)
- Single-threaded code paths
- Immutable data structures
- Read-only shared data

Output ONLY valid JSON:
{"findings":[{"id":"SEC-15-X","severity":"medium|high","file":"path","line":N,"title":"Race Condition","description":"[Variable/resource] accessed without synchronization between lines N and M","recommendation":"Add proper synchronization"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-16",
		Name:     "Information Disclosure",
		Category: CategorySecurity,
		CWE:      []string{"CWE-200", "CWE-532"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Information Disclosure (OWASP A01, CWE-200)

REPORT ONLY if you find these CONFIRMED patterns:
- Passwords/secrets logged: log.Print(password)
- Full stack traces returned in HTTP responses
- Debug mode enabled in production config
- Internal paths/IPs exposed in errors

DO NOT REPORT:
- Logging of non-sensitive data
- Error messages without sensitive details
- Debug logging behind feature flags
- Internal logging for troubleshooting

Output ONLY valid JSON:
{"findings":[{"id":"SEC-16-X","severity":"medium|high","file":"path","line":N,"title":"Information Disclosure","description":"[Sensitive data type] exposed via [method]","recommendation":"Remove sensitive data from [logs/responses]"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-17",
		Name:     "Insecure File Upload",
		Category: CategorySecurity,
		CWE:      []string{"CWE-434", "CWE-73"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Insecure File Operations (OWASP A04, CWE-434)

REPORT ONLY if you find these CONFIRMED patterns:
- File upload without type validation
- Uploaded files saved with user-provided names
- Files written with excessive permissions (0777)
- Temp files in predictable locations without randomization

DO NOT REPORT:
- File uploads with proper validation
- Generated/sanitized filenames
- Appropriate file permissions
- Secure temp file creation

Output ONLY valid JSON:
{"findings":[{"id":"SEC-17-X","severity":"high","file":"path","line":N,"title":"Insecure File Operation","description":"[Specific issue with file handling]","recommendation":"[Specific fix]"}],"summary":{"count":N}}`,
	},

	{
		ID:        "SEC-18",
		Name:      "Memory Safety",
		Category:  CategorySecurity,
		CWE:       []string{"CWE-119", "CWE-120"},
		Languages: []string{"c", "cpp", "rust"},
		Enabled:   true,
		Prompt: StandardPromptRules + `SCAN: Memory Safety (CWE-119)

REPORT ONLY if you find these CONFIRMED patterns:
- Buffer operations without bounds checking
- Use after free patterns
- Unchecked array indexing with external input
- Integer overflow in size calculations

DO NOT REPORT:
- Safe standard library functions
- Bounds-checked operations
- Rust safe code (only report unsafe blocks)

Output ONLY valid JSON:
{"findings":[{"id":"SEC-18-X","severity":"critical","file":"path","line":N,"title":"Memory Safety Issue","description":"[Buffer/pointer] issue at line N","recommendation":"Add bounds checking"}],"summary":{"count":N}}`,
	},

	{
		ID:        "SEC-19",
		Name:      "Prototype Pollution",
		Category:  CategorySecurity,
		CWE:       []string{"CWE-1321"},
		Languages: []string{"javascript", "typescript"},
		Enabled:   true,
		Prompt: StandardPromptRules + `SCAN: Prototype Pollution (CWE-1321)

REPORT ONLY if you find these CONFIRMED patterns:
- Object merge with user input: _.merge(obj, userInput)
- Direct __proto__ access with user data
- Recursive object copy without prototype check
- Object.assign with untrusted nested objects

DO NOT REPORT:
- Merges with trusted internal data
- Object.assign with flat objects
- Libraries with prototype pollution protection

Output ONLY valid JSON:
{"findings":[{"id":"SEC-19-X","severity":"high","file":"path","line":N,"title":"Prototype Pollution","description":"[Function] merges user input without prototype protection","recommendation":"Use Object.create(null) or validate input structure"}],"summary":{"count":N}}`,
	},

	{
		ID:       "SEC-20",
		Name:     "Open Redirect",
		Category: CategorySecurity,
		CWE:      []string{"CWE-601"},
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Open Redirect (CWE-601)

REPORT ONLY if you find these CONFIRMED patterns:
- HTTP redirect with user-controlled URL
- Location header set from user input
- window.location = userInput without validation

DO NOT REPORT:
- Redirects to whitelisted domains
- Relative path redirects only
- Redirects to hardcoded URLs

Output ONLY valid JSON:
{"findings":[{"id":"SEC-20-X","severity":"medium","file":"path","line":N,"title":"Open Redirect","description":"Redirect to user-controlled URL [variable]","recommendation":"Validate redirect URL against whitelist"}],"summary":{"count":N}}`,
	},

	// ============================================
	// QUALITY SCANS (Objective, Measurable Issues)
	// ============================================

	{
		ID:       "QUAL-01",
		Name:     "High Complexity Functions",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: High Complexity Functions

REPORT ONLY functions with:
- Cyclomatic complexity > 15 (many branches/conditions)
- Nesting depth > 5 levels
- More than 100 lines of code

DO NOT REPORT:
- Functions under these thresholds
- Generated code or configuration
- Test files with many test cases

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-01-X","severity":"medium","file":"path","line":N,"title":"High Complexity: [function]","description":"Function has complexity [N] / nesting [N] / lines [N]","recommendation":"Consider breaking into smaller functions"}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-02",
		Name:     "Unreachable Code",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Unreachable/Dead Code

REPORT ONLY if you find:
- Code after unconditional return/break/continue
- Conditions that are always true/false
- Unused exported functions (no callers found)
- Variables assigned but never read

DO NOT REPORT:
- Unused private functions (may be used via reflection)
- Code disabled by feature flags
- Test utilities

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-02-X","severity":"low","file":"path","line":N,"title":"Unreachable Code","description":"[Code description] is never executed","recommendation":"Remove dead code"}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-03",
		Name:     "Error Handling Issues",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Error Handling Issues

REPORT ONLY if you find:
- Errors explicitly ignored: err, _ := ...
- Empty catch/except blocks
- Error returned but never checked by caller
- Panic/throw in library code (not main/handler)

DO NOT REPORT:
- Intentionally ignored errors with comment explaining why
- Errors handled appropriately
- Deferred cleanup that can't fail meaningfully

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-03-X","severity":"medium","file":"path","line":N,"title":"Ignored Error","description":"Error from [function] ignored at line N","recommendation":"Handle or explicitly document why ignored"}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-04",
		Name:     "Resource Leaks",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Resource Leaks

REPORT ONLY if you find:
- File/connection opened but never closed
- Missing defer for closeable resources
- HTTP response body not closed
- Goroutines started but never joined/cancelled (Go)

DO NOT REPORT:
- Resources properly closed with defer/finally
- Resources managed by framework
- Short-lived resources in request handlers (auto-cleaned)

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-04-X","severity":"high","file":"path","line":N,"title":"Resource Leak","description":"[Resource type] opened at line N never closed","recommendation":"Add defer [resource].Close()"}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-05",
		Name:     "Concurrency Bugs",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Concurrency Issues

REPORT ONLY if you find:
- Data race: shared variable written from multiple goroutines/threads without sync
- Deadlock pattern: lock ordering inconsistent
- Channel send on closed channel
- WaitGroup Add after Wait

DO NOT REPORT:
- Properly synchronized concurrent code
- Intentionally lock-free code with atomic operations
- Channel patterns that are correct

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-05-X","severity":"high","file":"path","line":N,"title":"Concurrency Bug","description":"[Variable] has data race between lines N and M","recommendation":"Add synchronization"}],"summary":{"count":N,"score":0-100}}`,
	},

	{
		ID:       "QUAL-06",
		Name:     "Dependency Vulnerabilities",
		Category: CategoryQuality,
		Enabled:  true,
		Prompt: StandardPromptRules + `SCAN: Dependency Issues

REPORT ONLY if you find:
- Dependencies with known CVEs (check version in go.mod/package.json)
- Deprecated packages still in use
- Significantly outdated major versions (2+ years old)

DO NOT REPORT:
- Minor version differences
- Dependencies without known vulnerabilities
- Internal/private packages

Output ONLY valid JSON:
{"findings":[{"id":"QUAL-06-X","severity":"high|medium","file":"path","line":N,"title":"Vulnerable Dependency","description":"[package@version] has [CVE/issue]","recommendation":"Upgrade to [safe version]"}],"summary":{"count":N,"score":0-100}}`,
	},

	// Disable subjective quality scans by default
	{
		ID:       "QUAL-07",
		Name:     "Code Duplication",
		Category: CategoryQuality,
		Enabled:  false, // Disabled - subjective
		Prompt:   ``,
	},

	{
		ID:       "QUAL-08",
		Name:     "Naming Conventions",
		Category: CategoryQuality,
		Enabled:  false, // Disabled - style preference
		Prompt:   ``,
	},

	{
		ID:       "QUAL-09",
		Name:     "Magic Numbers",
		Category: CategoryQuality,
		Enabled:  false, // Disabled - design choice
		Prompt:   ``,
	},

	{
		ID:       "QUAL-10",
		Name:     "Documentation",
		Category: CategoryQuality,
		Enabled:  false, // Disabled - style preference
		Prompt:   ``,
	},

	{
		ID:       "QUAL-11",
		Name:     "Test Coverage",
		Category: CategoryQuality,
		Enabled:  false, // Disabled - requires context
		Prompt:   ``,
	},

	{
		ID:       "QUAL-12",
		Name:     "Performance",
		Category: CategoryQuality,
		Enabled:  false, // Disabled - requires benchmarks
		Prompt:   ``,
	},

	{
		ID:       "QUAL-13",
		Name:     "API Design",
		Category: CategoryQuality,
		Enabled:  false, // Disabled - design choice
		Prompt:   ``,
	},

	// ============================================
	// LANGUAGE-SPECIFIC SCANS (Security-Focused)
	// ============================================

	{
		ID:        "LANG-GO",
		Name:      "Go Security Patterns",
		Category:  CategoryLanguage,
		Languages: []string{"go"},
		Enabled:   true,
		Prompt: StandardPromptRules + `SCAN: Go Security Patterns

REPORT ONLY these CONFIRMED issues:
- Context not passed to long-running operations (database, HTTP calls)
- Goroutine leak: started goroutine with no way to stop
- unsafe package usage without clear justification
- Deferred function in loop (resource accumulation)

DO NOT REPORT:
- Style preferences (error naming, etc.)
- Interface design opinions
- Package organization suggestions

Output ONLY valid JSON:
{"findings":[{"id":"LANG-GO-X","severity":"high|medium","file":"path","line":N,"title":"[Issue]","description":"[Specific issue with evidence]","recommendation":"[Specific fix]"}],"summary":{"count":N}}`,
	},

	{
		ID:        "LANG-JS",
		Name:      "JavaScript Security Patterns",
		Category:  CategoryLanguage,
		Languages: []string{"javascript", "typescript"},
		Enabled:   true,
		Prompt: StandardPromptRules + `SCAN: JavaScript/TypeScript Security Patterns

REPORT ONLY these CONFIRMED issues:
- eval() with any external/dynamic input
- innerHTML/outerHTML with user data
- new Function() with user input
- document.write() usage
- Disabled TypeScript strict checks on security-sensitive code

DO NOT REPORT:
- Code style preferences
- Framework-specific patterns (React, Vue)
- Build/tooling configuration

Output ONLY valid JSON:
{"findings":[{"id":"LANG-JS-X","severity":"high|medium","file":"path","line":N,"title":"[Issue]","description":"[Specific issue with evidence]","recommendation":"[Specific fix]"}],"summary":{"count":N}}`,
	},

	{
		ID:        "LANG-PY",
		Name:      "Python Security Patterns",
		Category:  CategoryLanguage,
		Languages: []string{"python"},
		Enabled:   true,
		Prompt: StandardPromptRules + `SCAN: Python Security Patterns

REPORT ONLY these CONFIRMED issues:
- pickle.loads() with untrusted data
- eval()/exec() with user input
- subprocess with shell=True and user input
- yaml.load() without Loader (unsafe default)
- assert statements for security checks (disabled with -O)

DO NOT REPORT:
- Type hint suggestions
- Style/formatting (PEP8)
- Import organization

Output ONLY valid JSON:
{"findings":[{"id":"LANG-PY-X","severity":"high|medium","file":"path","line":N,"title":"[Issue]","description":"[Specific issue with evidence]","recommendation":"[Specific fix]"}],"summary":{"count":N}}`,
	},

	{
		ID:        "LANG-JAVA",
		Name:      "Java Security Patterns",
		Category:  CategoryLanguage,
		Languages: []string{"java"},
		Enabled:   true,
		Prompt: StandardPromptRules + `SCAN: Java Security Patterns

REPORT ONLY these CONFIRMED issues:
- ObjectInputStream with untrusted data
- Runtime.exec() with user input
- XML parsing without disabling external entities
- Insecure random (java.util.Random for security)
- SQL string concatenation

DO NOT REPORT:
- Code style preferences
- Null safety suggestions (unless causing bugs)
- Generic best practices

Output ONLY valid JSON:
{"findings":[{"id":"LANG-JAVA-X","severity":"high|medium","file":"path","line":N,"title":"[Issue]","description":"[Specific issue with evidence]","recommendation":"[Specific fix]"}],"summary":{"count":N}}`,
	},

	{
		ID:        "LANG-RUST",
		Name:      "Rust Security Patterns",
		Category:  CategoryLanguage,
		Languages: []string{"rust"},
		Enabled:   true,
		Prompt: StandardPromptRules + `SCAN: Rust Security Patterns

REPORT ONLY these CONFIRMED issues:
- unsafe blocks without safety comments/justification
- .unwrap() on user input (can panic)
- Raw pointer dereferencing without bounds check
- std::mem::transmute usage

DO NOT REPORT:
- .unwrap() in tests or with known-safe values
- unsafe in FFI bindings (expected)
- Clippy-style suggestions

Output ONLY valid JSON:
{"findings":[{"id":"LANG-RUST-X","severity":"high|medium","file":"path","line":N,"title":"[Issue]","description":"[Specific issue with evidence]","recommendation":"[Specific fix]"}],"summary":{"count":N}}`,
	},

	{
		ID:        "LANG-C",
		Name:      "C/C++ Security Patterns",
		Category:  CategoryLanguage,
		Languages: []string{"c", "cpp"},
		Enabled:   true,
		Prompt: StandardPromptRules + `SCAN: C/C++ Security Patterns

REPORT ONLY these CONFIRMED issues:
- strcpy/sprintf without bounds (use strncpy/snprintf)
- gets() usage (always unsafe)
- Buffer size from user input without validation
- Format string with user input: printf(userInput)
- Unchecked return from malloc

DO NOT REPORT:
- Modern C++ safe equivalents in use
- Static analysis tool annotations
- Coding style preferences

Output ONLY valid JSON:
{"findings":[{"id":"LANG-C-X","severity":"critical|high","file":"path","line":N,"title":"[Issue]","description":"[Specific issue with evidence]","recommendation":"[Specific fix]"}],"summary":{"count":N}}`,
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
