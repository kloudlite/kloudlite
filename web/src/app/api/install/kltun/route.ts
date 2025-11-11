import { NextRequest, NextResponse } from 'next/server';

const INSTALL_SCRIPT = `#!/bin/bash
set -e

# Colors
RED='\\033[0;31m'
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
BLUE='\\033[0;34m'
NC='\\033[0m' # No Color

# Base URL for downloads - will be replaced with actual subdomain
DOWNLOAD_BASE_URL="\${KLTUN_SERVER_URL:-https://KLTUN_BASE_URL/api/download/kltun}"

# Detect OS and Architecture
detect_platform() {
    local os=""
    local arch=""

    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        MINGW*|MSYS*|CYGWIN*) os="windows";;
        *)          echo -e "\${RED}Error: Unsupported operating system\${NC}"; exit 1;;
    esac

    # Detect Architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64";;
        arm64|aarch64)  arch="arm64";;
        *)              echo -e "\${RED}Error: Unsupported architecture: $(uname -m)\${NC}"; exit 1;;
    esac

    echo "\${os}-\${arch}"
}

# Download kltun binary
download_kltun() {
    local platform=$1
    local version=\${2:-latest}
    local download_url="\${DOWNLOAD_BASE_URL}/\${platform}"

    if [ "\$version" != "latest" ]; then
        download_url="\${download_url}?version=\${version}"
    fi

    echo -e "\${BLUE}Downloading kltun for \${platform}...\${NC}"

    if command -v curl &> /dev/null; then
        curl -fsSL "\${download_url}" -o kltun
    elif command -v wget &> /dev/null; then
        wget -q "\${download_url}" -O kltun
    else
        echo -e "\${RED}Error: Neither curl nor wget found. Please install one of them.\${NC}"
        exit 1
    fi

    chmod +x kltun
    echo -e "\${GREEN}✓ Downloaded successfully\${NC}"
}

# Install kltun to system
install_kltun() {
    local install_dir="/usr/local/bin"

    if [ -w "\$install_dir" ]; then
        mv kltun "\${install_dir}/kltun"
        echo -e "\${GREEN}✓ Installed to \${install_dir}/kltun\${NC}"
    else
        echo -e "\${YELLOW}Installing to \${install_dir} requires sudo privileges\${NC}"
        sudo mv kltun "\${install_dir}/kltun"
        echo -e "\${GREEN}✓ Installed to \${install_dir}/kltun\${NC}"
    fi
}

# Verify installation
verify_installation() {
    if command -v kltun &> /dev/null; then
        echo -e "\${GREEN}✓ kltun installed successfully!\${NC}"
        echo -e "\${BLUE}Version: \${NC}$(kltun version 2>/dev/null || echo 'unknown')"
        return 0
    else
        echo -e "\${RED}✗ Installation failed\${NC}"
        return 1
    fi
}

# Print usage
print_usage() {
    cat << EOF
\${BLUE}Kloudlite Tunnel (kltun) - Installation Script\${NC}

Usage:
  curl -fsSL https://KLTUN_BASE_URL/kltun | bash
  curl -fsSL https://KLTUN_BASE_URL/kltun | bash -s -- [OPTIONS]

Options:
  --version VERSION    Install specific version (default: latest)
  --help              Show this help message

Examples:
  # Install latest version
  curl -fsSL https://KLTUN_BASE_URL/kltun | bash

  # Install specific version
  curl -fsSL https://KLTUN_BASE_URL/kltun | bash -s -- --version 0.1.0

After installation:
  kltun --help                Show kltun help
  kltun connect               Connect to Kloudlite VPN
  kltun install-ca            Install CA certificates
  kltun hosts add             Manage hosts file entries

EOF
}

# Main installation flow
main() {
    local version="latest"

    # Parse arguments
    while [ $# -gt 0 ]; do
        case $1 in
            --version)
                version="$2"
                shift 2
                ;;
            --help)
                print_usage
                exit 0
                ;;
            *)
                echo -e "\${RED}Unknown option: $1\${NC}"
                print_usage
                exit 1
                ;;
        esac
    done

    echo -e "\${BLUE}================================\${NC}"
    echo -e "\${BLUE}  Kloudlite Tunnel (kltun)\${NC}"
    echo -e "\${BLUE}================================\${NC}"
    echo ""

    # Detect platform
    platform=$(detect_platform)
    echo -e "\${BLUE}Detected platform: \${platform}\${NC}"
    echo ""

    # Download
    download_kltun "\$platform" "\$version"
    echo ""

    # Install
    install_kltun
    echo ""

    # Verify
    if verify_installation; then
        echo ""
        echo -e "\${GREEN}🎉 Installation complete!\${NC}"
        echo ""
        echo -e "\${BLUE}Quick start:\${NC}"
        echo "  kltun --help          # Show help"
        echo "  kltun daemon          # Start daemon"
        echo "  kltun connect         # Connect to VPN"
        echo ""
        echo -e "\${BLUE}Documentation:\${NC}"
        echo "  https://github.com/kloudlite/kloudlite/tree/development/api/cmd/kltun"
    else
        exit 1
    fi
}

main "$@"
`;

export async function GET(request: NextRequest) {
  // Get the base URL from the request
  const protocol = request.headers.get('x-forwarded-proto') || 'https';
  const host = request.headers.get('host') || request.nextUrl.host;
  const baseUrl = `${protocol}://${host}`;

  // Replace placeholder with actual base URL
  const customizedScript = INSTALL_SCRIPT.replace(/KLTUN_BASE_URL/g, baseUrl);

  return new NextResponse(customizedScript, {
    headers: {
      'Content-Type': 'text/x-shellscript',
      'Content-Disposition': 'inline; filename="install.sh"',
      'Cache-Control': 'public, max-age=300', // Cache for 5 minutes
    },
  });
}

export const runtime = 'edge';
