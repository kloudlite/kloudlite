# kltun - Kloudlite Tunnel Manager

`kltun` is a CLI tool for managing secure tunnels to Kloudlite workspaces and services. It provides comprehensive CA certificate management across multiple platforms and trust stores, and embeds a proxyguard client for secure UDP-to-TCP tunneling.

## Features

- **Secure Tunnel Connection**: Embedded proxyguard client for UDP-to-TCP proxy tunneling (WireGuard support)
- **Cross-Platform CA Installation**: Automatically installs CA certificates to system trust stores on macOS, Linux, and Windows
- **Multi-Trust-Store Support**: Supports system stores, NSS/Firefox, and Java keystores
- **Intelligent Privilege Management**: Automatically handles sudo/administrator privileges when needed
- **Graceful Degradation**: Continues working even if some trust stores are unavailable
- **Comprehensive Error Handling**: Provides detailed error messages and installation guides
- **TLS 1.3 Support**: Secure HTTPS connections to Kloudlite servers

## Installation

Build from source:

```bash
go build -o kltun ./cmd/kltun
```

Or use the pre-built binary from releases.

## Usage

### Connect to Kloudlite Workspace

Establish a secure tunnel to a Kloudlite workspace:

```bash
kltun connect --server https://workspace.kloudlite.io
```

With custom ports:

```bash
kltun connect --server https://workspace.kloudlite.io --listen-port 51821 --forward-port 51820
```

With fwmark on Linux (requires root):

```bash
sudo kltun connect --server https://workspace.kloudlite.io --fwmark 51820
```

With pre-resolved IPs (useful for boot-time scenarios):

```bash
kltun connect --server https://workspace.kloudlite.io --peer-ips 203.0.113.1,203.0.113.2
```

The tunnel will:
- Listen on local UDP port 51821 (configurable with `--listen-port`)
- Forward traffic from WireGuard port 51820 (configurable with `--forward-port`)
- Convert UDP to TCP/HTTPS and proxy to the Kloudlite server
- Use TLS 1.3 for secure communication
- Auto-reconnect with exponential backoff on connection failures

**Configure WireGuard to use the tunnel:**

```ini
[Interface]
PrivateKey = ...
Address = ...
DNS = ...
ListenPort = 51820

[Peer]
PublicKey = ...
AllowedIPs = ...
Endpoint = 127.0.0.1:51821
```

### Install CA Certificate

Install the Kloudlite CA certificate to all available trust stores:

```bash
kltun install-ca --cert /path/to/ca.crt
```

Install to specific trust stores only:

```bash
# System trust store only
kltun install-ca --cert /path/to/ca.crt --stores system

# System and Firefox
kltun install-ca --cert /path/to/ca.crt --stores system,nss

# All stores explicitly
kltun install-ca --cert /path/to/ca.crt --stores system,nss,java
```

### Uninstall CA Certificate

Remove the CA certificate from all trust stores:

```bash
kltun uninstall-ca --cert /path/to/ca.crt
```

Remove from specific trust stores:

```bash
kltun uninstall-ca --cert /path/to/ca.crt --stores system,nss
```

### Version Information

```bash
kltun version
```

## Supported Trust Stores

### System Trust Store

- **macOS**: System Keychain (`/Library/Keychains/System.keychain`)
- **Linux**: Distribution-specific certificate managers
  - Red Hat/Fedora/CentOS: `/etc/pki/ca-trust/source/anchors/`
  - Debian/Ubuntu: `/usr/local/share/ca-certificates/`
  - Arch Linux: `/etc/ca-certificates/trust-source/anchors/`
  - openSUSE: `/usr/share/pki/trust/anchors`
- **Windows**: Windows Certificate Store (ROOT)

### NSS/Firefox Trust Store

Automatically detects and installs to:
- Standard NSS databases (`~/.pki/nssdb`)
- Firefox profiles (all profiles automatically detected)
- Snap-packaged browsers (Chromium, Firefox)
- System-wide NSS databases (`/etc/pki/nssdb`)

### Java Trust Store

Installs to Java's `cacerts` keystore when `JAVA_HOME` is set:
- Modern Java (9+): `$JAVA_HOME/lib/security/cacerts`
- Older Java: `$JAVA_HOME/jre/lib/security/cacerts`

## Requirements

### Base Requirements

- Go 1.18+ (for building from source)
- Sudo/administrator privileges (for system-level operations)

### Platform-Specific Requirements

#### macOS
- `security` command (built-in)
- `brew` (optional, for NSS tools)

#### Linux
- Distribution-specific certificate management tools (usually pre-installed):
  - `update-ca-trust` (Red Hat/Fedora/CentOS)
  - `update-ca-certificates` (Debian/Ubuntu/openSUSE)
  - `trust` (Arch Linux)

#### Windows
- No additional requirements (uses native APIs)

### Optional Requirements

#### For NSS/Firefox Support
- **macOS**: `brew install nss`
- **Debian/Ubuntu**: `apt install libnss3-tools`
- **Red Hat/CentOS**: `yum install nss-tools`
- **Arch Linux**: `pacman -S nss`
- **openSUSE**: `zypper install mozilla-nss-tools`

#### For Java Support
- `JAVA_HOME` environment variable set
- JDK or JRE installation

## How It Works

### macOS Implementation

1. Adds certificate to System Keychain using `security add-trusted-cert`
2. Exports trust settings to temporary plist file
3. Modifies trust settings to explicitly trust for SSL and X.509
4. Imports modified trust settings back

### Linux Implementation

1. Detects distribution by checking filesystem paths
2. Copies certificate to appropriate system trust directory
3. Runs distribution-specific update command to refresh trust store

### Windows Implementation

1. Opens ROOT certificate store using Windows API
2. Adds certificate using `CertAddEncodedCertificateToStore`
3. Verifies installation by checking certificate presence

### NSS/Firefox Implementation

1. Finds `certutil` binary (checks PATH and Homebrew on macOS)
2. Detects all NSS databases and Firefox profiles
3. Installs to each profile using `certutil -A` with trust flags `C,,`
4. Automatically retries with sudo on permission errors

### Java Implementation

1. Checks `JAVA_HOME` environment variable
2. Finds `keytool` and `cacerts` in Java installation
3. Imports certificate using default keystore password (`changeit`)
4. Automatically retries with sudo on permission errors

## Edge Cases Handled

### Permission Management
- Automatically detects if already running as root
- Uses sudo only when necessary
- Custom password prompts for clarity
- Graceful fallback if sudo unavailable

### Platform Detection
- Filesystem-based Linux distribution detection
- Dual cacerts path support for old and new Java versions
- Both SQL and DBM NSS database formats
- Snap package support for browsers

### Error Handling
- Continues on non-critical failures
- Provides detailed error messages with context
- Suggests installation commands for missing tools
- Graceful degradation when trust stores unavailable

### Certificate Management
- Removes legacy certificate filenames on uninstall
- Handles multiple Firefox profiles automatically
- Verifies installation success before completing
- Safe certificate enumeration on Windows

## Architecture

```
kltun/
├── main.go                          # Entry point
├── cmd/
│   ├── root.go                      # Root command
│   ├── install_ca.go                # Install command
│   └── uninstall_ca.go              # Uninstall command
└── pkg/
    └── truststore/
        ├── truststore.go            # Common interface
        ├── truststore_darwin.go     # macOS (build tag: darwin)
        ├── truststore_linux.go      # Linux (build tag: linux)
        ├── truststore_windows.go    # Windows (build tag: windows)
        ├── truststore_nss.go        # NSS/Firefox (all platforms)
        └── truststore_java.go       # Java keystore (all platforms)
```

## Security Considerations

### Privilege Escalation
- Minimal scope: Only specific commands elevated
- Transparent prompting with custom sudo messages
- No password storage or caching
- Automatic root detection to avoid unnecessary elevation

### Certificate Security
- Validates certificate format before installation
- Uses secure file permissions (0600 for temporary files)
- Cleans up temporary files after use
- Only installs to trusted system locations

### Trust Store Integrity
- Explicit trust policies on macOS (SSL and X.509)
- Serial number matching on Windows (prevents incorrect removal)
- Unique certificate aliases to avoid conflicts
- Verification after installation

## Troubleshooting

### Certificate not trusted after installation
- Restart your browser
- Check if your application trusts the system certificate store
- Verify installation with: `security find-certificate -c "kloudlite-ca"` (macOS)

### Permission denied errors
- Ensure sudo is available and configured
- Check if you have administrator privileges
- Try running with explicit sudo: `sudo kltun install-ca ...`

### NSS tools not found
- Install appropriate package for your system (see Requirements)
- Verify `certutil` is in PATH: `which certutil`
- On macOS, ensure Homebrew NSS is installed: `brew install nss`

### Java keystore errors
- Ensure `JAVA_HOME` is set: `echo $JAVA_HOME`
- Verify keytool exists: `$JAVA_HOME/bin/keytool -version`
- Check cacerts file permissions

## Contributing

When adding new features or fixing bugs:

1. Follow the existing code structure
2. Add platform-specific code to appropriate `truststore_*.go` files
3. Update this README with new features or requirements
4. Test on target platform before submitting PR

## License

[Add your license here]

## Credits

Based on research from [mkcert](https://github.com/FiloSottile/mkcert) by Filippo Valsorda.
