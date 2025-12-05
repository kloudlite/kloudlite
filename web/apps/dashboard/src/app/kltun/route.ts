import { NextRequest, NextResponse } from 'next/server'

/**
 * GET /kltun
 *
 * Returns an idempotent installation script for kltun
 * Usage: curl -fsSL https://example.com/kltun | sh -s -- --token <TOKEN> [--server <SERVER>]
 */
export async function GET(request: NextRequest) {
  // Get the server URL from the request
  const protocol = request.headers.get('x-forwarded-proto') || 'https'
  const host = request.headers.get('host') || request.nextUrl.host
  const serverUrl = `${protocol}://${host}`
  const downloadBaseUrl = `${serverUrl}/api/download/kltun`

  const script = `#!/bin/sh
# kltun Installation and Connection Script
# Supports: Linux, macOS, Windows (Git Bash/WSL/MSYS2)

set -e

# Colors for output
RED='\\033[0;31m'
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
BLUE='\\033[0;34m'
NC='\\033[0m' # No Color

# Portable echo function
echo_color() { printf "%b\n" "$1"; }

# Default values
TOKEN=""
SERVER="${serverUrl}"
PLATFORM=""
ARCH=""

# Parse arguments
while [ $# -gt 0 ]; do
    case $1 in
        --token)
            TOKEN="$2"
            shift 2
            ;;
        --platform)
            PLATFORM="$2"
            shift 2
            ;;
        --arch)
            ARCH="$2"
            shift 2
            ;;
        *)
            echo_color "\${RED}Unknown option: $1\${NC}"
            echo "Usage: curl -fsSL ${serverUrl}/kltun | sh -s -- --token <TOKEN>"
            exit 1
            ;;
    esac
done

# Validate required parameters
if [ -z "$TOKEN" ]; then
    echo_color "\${RED}Error: --token is required\${NC}"
    echo "Usage: curl -fsSL ${serverUrl}/kltun | sh -s -- --token <TOKEN>"
    exit 1
fi

# Detect platform if not specified
if [ -z "$PLATFORM" ]; then
    case "$(uname -s)" in
        Linux*)     PLATFORM="linux";;
        Darwin*)    PLATFORM="darwin";;
        MINGW*|MSYS*|CYGWIN*) PLATFORM="windows";;
        *)
            echo_color "\${RED}Error: Unsupported operating system: $(uname -s)\${NC}"
            echo "Supported: Linux, macOS, Windows (Git Bash/WSL/MSYS2)"
            exit 1
            ;;
    esac
fi

# Detect architecture if not specified
if [ -z "$ARCH" ]; then
    case "$(uname -m)" in
        x86_64|amd64)   ARCH="amd64";;
        arm64|aarch64)  ARCH="arm64";;
        *)
            echo_color "\${RED}Error: Unsupported architecture: $(uname -m)\${NC}"
            echo "Supported: x86_64/amd64, arm64/aarch64"
            exit 1
            ;;
    esac
fi

echo_color "\${BLUE}================================\${NC}"
echo_color "\${BLUE}  kltun Installation\${NC}"
echo_color "\${BLUE}================================\${NC}"
echo ""
echo_color "\${BLUE}Platform:\${NC} \${PLATFORM}-\${ARCH}"
echo_color "\${BLUE}Server:\${NC} \${SERVER}"
echo ""

# Download URL
DOWNLOAD_URL="${downloadBaseUrl}/\${PLATFORM}-\${ARCH}"
SHA_URL="${downloadBaseUrl}/\${PLATFORM}-\${ARCH}/sha"

# Check if kltun is already installed and up-to-date
NEEDS_UPDATE=true
if [ "$PLATFORM" = "windows" ]; then
    KLTUN_BIN="/c/Program Files/kltun/kltun.exe"
else
    KLTUN_BIN="/usr/local/bin/kltun"
fi

if [ -f "$KLTUN_BIN" ]; then
    echo_color "\${BLUE}Checking installed kltun version...\${NC}"
    # Get remote SHA
    REMOTE_SHA_JSON=$(curl -fsSL "$SHA_URL" 2>/dev/null || echo "")
    if [ -n "$REMOTE_SHA_JSON" ]; then
        REMOTE_SHA=$(echo "$REMOTE_SHA_JSON" | grep -o '"sha256":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$REMOTE_SHA" ]; then
            # Calculate local SHA
            if command -v sha256sum >/dev/null 2>&1; then
                LOCAL_SHA=$(sha256sum "$KLTUN_BIN" 2>/dev/null | awk '{print \$1}')
            elif command -v shasum >/dev/null 2>&1; then
                LOCAL_SHA=$(shasum -a 256 "$KLTUN_BIN" 2>/dev/null | awk '{print \$1}')
            fi

            if [ "$LOCAL_SHA" = "$REMOTE_SHA" ]; then
                echo_color "\${GREEN}✓ kltun is already up-to-date (SHA: \${LOCAL_SHA:0:12}...)\${NC}"
                NEEDS_UPDATE=false
            else
                echo_color "\${YELLOW}kltun update available\${NC}"
                echo_color "\${BLUE}  Local:  \${LOCAL_SHA:0:12}...\${NC}"
                echo_color "\${BLUE}  Remote: \${REMOTE_SHA:0:12}...\${NC}"
            fi
        fi
    fi
fi

# Platform-specific installation
if [ "$NEEDS_UPDATE" = true ] && [ "$PLATFORM" = "windows" ]; then
    # Windows installation
    echo_color "\${BLUE}Installing on Windows...\${NC}"

    INSTALL_DIR="/c/Program Files/kltun"
    TEMP_FILE="/tmp/kltun.exe"

    echo_color "\${BLUE}Downloading kltun...\${NC}"
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"; then
        echo_color "\${RED}Error: Failed to download kltun\${NC}"
        exit 1
    fi

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Move to install directory
    echo_color "\${BLUE}Installing to $INSTALL_DIR...\${NC}"
    mv -f "$TEMP_FILE" "$INSTALL_DIR/kltun.exe"

    # Add to PATH if not already present
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        echo_color "\${YELLOW}Note: Add $INSTALL_DIR to your PATH to use kltun from anywhere\${NC}"
    fi

    KLTUN_CMD="$INSTALL_DIR/kltun.exe"

elif [ "$NEEDS_UPDATE" = true ]; then
    # Unix-like systems (Linux, macOS)
    echo_color "\${BLUE}Installing on Unix-like system...\${NC}"

    # Detect if we need sudo
    SUDO=""
    if [ "$(id -u)" != "0" ]; then
        if [ -w "/usr/local/bin" ]; then
            SUDO=""
        else
            SUDO="sudo"
        fi
    fi

    # Download kltun to temp location
    TEMP_FILE="/tmp/kltun-$$.tmp"
    echo_color "\${BLUE}Downloading kltun...\${NC}"
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"; then
        echo_color "\${RED}Error: Failed to download kltun\${NC}"
        exit 1
    fi

    # Make executable
    chmod +x "$TEMP_FILE"

    # Install to /usr/local/bin
    INSTALL_PATH="/usr/local/bin/kltun"
    if [ -f "$INSTALL_PATH" ]; then
        echo_color "\${YELLOW}kltun already installed, updating...\${NC}"
        if [ -n "$SUDO" ]; then
            $SUDO mv -f "$TEMP_FILE" "$INSTALL_PATH"
        else
            mv -f "$TEMP_FILE" "$INSTALL_PATH"
        fi
    else
        echo_color "\${BLUE}Installing to $INSTALL_PATH...\${NC}"
        if [ -n "$SUDO" ]; then
            $SUDO mv "$TEMP_FILE" "$INSTALL_PATH"
        else
            mv "$TEMP_FILE" "$INSTALL_PATH"
        fi
    fi

    KLTUN_CMD="kltun"
else
    # No update needed - set command path
    if [ "$PLATFORM" = "windows" ]; then
        KLTUN_CMD="/c/Program Files/kltun/kltun.exe"
    else
        KLTUN_CMD="kltun"
    fi
fi

if [ "$NEEDS_UPDATE" = true ]; then
    echo_color "\${GREEN}✓ kltun installed successfully!\${NC}"

    # Stop daemon if running (to ensure new binary is used)
    echo_color "\${BLUE}Restarting daemon with new binary...\${NC}"
    if [ "$PLATFORM" = "darwin" ]; then
        # macOS: use launchctl to restart the launchd-managed daemon (system daemons require sudo)
        if sudo launchctl list 2>/dev/null | grep -q "io.kloudlite.kltund"; then
            echo_color "\${BLUE}Restarting launchd daemon...\${NC}"
            sudo launchctl kickstart -k system/io.kloudlite.kltund 2>/dev/null || true
            sleep 2
        fi
    elif [ "$PLATFORM" = "windows" ]; then
        if $KLTUN_CMD daemon status >/dev/null 2>&1; then
            taskkill //F //IM kltun.exe 2>/dev/null || true
            sleep 1
        fi
    else
        # Linux: use pkill
        if $KLTUN_CMD daemon status >/dev/null 2>&1; then
            if [ "$(id -u)" != "0" ] && [ ! -w "/usr/local/bin" ]; then
                sudo pkill -9 kltun 2>/dev/null || true
            else
                pkill -9 kltun 2>/dev/null || true
            fi
            sleep 1
        fi
    fi
    echo ""
fi
echo ""

# Start daemon if not running
echo_color "\${BLUE}Starting kltun daemon...\${NC}"
if [ "$PLATFORM" = "darwin" ]; then
    # macOS: check if launchd daemon is running (system daemons require sudo)
    if sudo launchctl list 2>/dev/null | grep -q "io.kloudlite.kltund"; then
        echo_color "\${GREEN}✓ Daemon running (launchd)\${NC}"
    else
        # Launchd plist not installed, install it first
        echo_color "\${BLUE}Installing launchd daemon...\${NC}"
        sudo $KLTUN_CMD daemon install 2>/dev/null || true
        sleep 2
        if sudo launchctl list 2>/dev/null | grep -q "io.kloudlite.kltund"; then
            echo_color "\${GREEN}✓ Daemon installed and running\${NC}"
        else
            echo_color "\${RED}Error: Failed to install daemon\${NC}"
            echo_color "\${YELLOW}Try installing it manually:\${NC}"
            echo "  sudo $KLTUN_CMD daemon install"
            exit 1
        fi
    fi
elif [ "$PLATFORM" = "windows" ]; then
    # Windows: start daemon in background
    if ! $KLTUN_CMD daemon status >/dev/null 2>&1; then
        start /B $KLTUN_CMD daemon run
        sleep 3
    fi
else
    # Linux: start daemon in background
    if ! $KLTUN_CMD daemon status >/dev/null 2>&1; then
        nohup $KLTUN_CMD daemon run >/dev/null 2>&1 &

        # Wait for daemon to start (max 5 seconds)
        for i in 1 2 3 4 5; do
            sleep 1
            if $KLTUN_CMD daemon status >/dev/null 2>&1; then
                echo_color "\${GREEN}✓ Daemon started\${NC}"
                break
            fi
        done

        # Final check
        if ! $KLTUN_CMD daemon status >/dev/null 2>&1; then
            echo_color "\${RED}Error: Failed to start daemon\${NC}"
            echo_color "\${YELLOW}Try starting it manually:\${NC}"
            echo "  $KLTUN_CMD daemon run"
            exit 1
        fi
    else
        echo_color "\${GREEN}✓ Daemon already running\${NC}"
    fi
fi

echo_color "\${BLUE}Connecting to VPN...\${NC}"

# Connect to VPN
if ! $KLTUN_CMD connect --token "$TOKEN" --server "$SERVER"; then
    echo_color "\${RED}Error: Failed to connect to VPN\${NC}"
    echo_color "\${YELLOW}You can retry manually with:\${NC}"
    echo "  $KLTUN_CMD daemon run  # Start daemon first"
    echo "  $KLTUN_CMD connect --token $TOKEN --server $SERVER"
    exit 1
fi

# Verify connection is actually established (not just initiated)
echo_color "\${BLUE}Verifying VPN connection...\${NC}"
TIMEOUT=60
ELAPSED=0
CONNECTED=false

while [ $ELAPSED -lt $TIMEOUT ]; do
    # Check if connection is active by pinging the VPN gateway
    if ping -c 1 -W 1 10.17.0.1 >/dev/null 2>&1; then
        CONNECTED=true
        break
    fi

    printf "\${BLUE}.\${NC}"
    sleep 2
    ELAPSED=\$((ELAPSED + 2))
done
echo ""

if [ "$CONNECTED" = false ]; then
    echo_color "\${RED}Error: VPN connection timeout after \${TIMEOUT} seconds\${NC}"
    echo_color "\${YELLOW}Check the logs for details:\${NC}"
    echo "  tail -f /var/log/kltund.log"
    echo ""
    echo_color "\${YELLOW}Common issues:\${NC}"
    echo "  - VPN server may be down or unreachable"
    echo "  - Invalid token"
    echo "  - Network connectivity issues"
    echo "  - WireGuard tunnel not fully established"
    exit 1
fi

echo ""
echo_color "\${GREEN}✓ Connected successfully!\${NC}"
`

  return new NextResponse(script, {
    status: 200,
    headers: {
      'Content-Type': 'text/x-shellscript',
      'Content-Disposition': 'inline; filename="install-kltun.sh"',
      'Cache-Control': 'no-cache',
    },
  })
}
