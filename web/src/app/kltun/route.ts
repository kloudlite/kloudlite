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
            echo -e "\${RED}Unknown option: $1\${NC}"
            echo "Usage: curl -fsSL ${serverUrl}/kltun | sh -s -- --token <TOKEN>"
            exit 1
            ;;
    esac
done

# Validate required parameters
if [ -z "$TOKEN" ]; then
    echo -e "\${RED}Error: --token is required\${NC}"
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
            echo -e "\${RED}Error: Unsupported operating system: $(uname -s)\${NC}"
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
            echo -e "\${RED}Error: Unsupported architecture: $(uname -m)\${NC}"
            echo "Supported: x86_64/amd64, arm64/aarch64"
            exit 1
            ;;
    esac
fi

echo -e "\${BLUE}================================\${NC}"
echo -e "\${BLUE}  kltun Installation\${NC}"
echo -e "\${BLUE}================================\${NC}"
echo ""
echo -e "\${BLUE}Platform:\${NC} \${PLATFORM}-\${ARCH}"
echo -e "\${BLUE}Server:\${NC} \${SERVER}"
echo ""

# Download URL
DOWNLOAD_URL="${downloadBaseUrl}/\${PLATFORM}-\${ARCH}"

# Platform-specific installation
if [ "$PLATFORM" = "windows" ]; then
    # Windows installation
    echo -e "\${BLUE}Installing on Windows...\${NC}"

    INSTALL_DIR="/c/Program Files/kltun"
    TEMP_FILE="/tmp/kltun.exe"

    echo -e "\${BLUE}Downloading kltun...\${NC}"
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"; then
        echo -e "\${RED}Error: Failed to download kltun\${NC}"
        exit 1
    fi

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Move to install directory
    echo -e "\${BLUE}Installing to $INSTALL_DIR...\${NC}"
    mv -f "$TEMP_FILE" "$INSTALL_DIR/kltun.exe"

    # Add to PATH if not already present
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        echo -e "\${YELLOW}Note: Add $INSTALL_DIR to your PATH to use kltun from anywhere\${NC}"
    fi

    KLTUN_CMD="$INSTALL_DIR/kltun.exe"

else
    # Unix-like systems (Linux, macOS)
    echo -e "\${BLUE}Installing on Unix-like system...\${NC}"

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
    echo -e "\${BLUE}Downloading kltun...\${NC}"
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"; then
        echo -e "\${RED}Error: Failed to download kltun\${NC}"
        exit 1
    fi

    # Make executable
    chmod +x "$TEMP_FILE"

    # Install to /usr/local/bin
    INSTALL_PATH="/usr/local/bin/kltun"
    if [ -f "$INSTALL_PATH" ]; then
        echo -e "\${YELLOW}kltun already installed, updating...\${NC}"
        if [ -n "$SUDO" ]; then
            $SUDO mv -f "$TEMP_FILE" "$INSTALL_PATH"
        else
            mv -f "$TEMP_FILE" "$INSTALL_PATH"
        fi
    else
        echo -e "\${BLUE}Installing to $INSTALL_PATH...\${NC}"
        if [ -n "$SUDO" ]; then
            $SUDO mv "$TEMP_FILE" "$INSTALL_PATH"
        else
            mv "$TEMP_FILE" "$INSTALL_PATH"
        fi
    fi

    KLTUN_CMD="kltun"
fi

echo -e "\${GREEN}✓ kltun installed successfully!\${NC}"
echo ""
echo -e "\${BLUE}Connecting to VPN...\${NC}"

# Connect to VPN
if ! $KLTUN_CMD connect --token "$TOKEN" --server "$SERVER"; then
    echo -e "\${RED}Error: Failed to connect to VPN\${NC}"
    echo -e "\${YELLOW}You can retry manually with:\${NC}"
    echo "  $KLTUN_CMD connect --token $TOKEN --server $SERVER"
    exit 1
fi

echo ""
echo -e "\${GREEN}✓ Connected successfully!\${NC}"
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
