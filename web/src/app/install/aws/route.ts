import { NextRequest, NextResponse } from 'next/server';

const INSTALL_SCRIPT = `#!/bin/bash
set -e

# Colors
RED='\\033[0;31m'
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
BLUE='\\033[0;34m'
NC='\\033[0m' # No Color

# Base URL for downloads
DOWNLOAD_BASE_URL="https://get.khost.dev/api/download/kli"

# Parse installation key from arguments
INSTALLATION_KEY=""
AWS_REGION=""

while [ $# -gt 0 ]; do
    case $1 in
        --key)
            INSTALLATION_KEY="$2"
            shift 2
            ;;
        --region)
            AWS_REGION="$2"
            shift 2
            ;;
        *)
            echo -e "\${RED}Unknown option: $1\${NC}"
            exit 1
            ;;
    esac
done

if [ -z "\$INSTALLATION_KEY" ]; then
    echo -e "\${RED}Error: --key parameter is required\${NC}"
    echo "Usage: curl -fsSL https://get.khost.dev/install/aws | bash -s -- --key YOUR_KEY [--region REGION]"
    exit 1
fi

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

# Download kli binary
download_kli() {
    local platform=$1
    local download_url="\${DOWNLOAD_BASE_URL}/\${platform}"

    echo -e "\${BLUE}Downloading kli for \${platform}...\${NC}"

    if command -v curl &> /dev/null; then
        curl -fsSL "\${download_url}" -o kli
    elif command -v wget &> /dev/null; then
        wget -q "\${download_url}" -O kli
    else
        echo -e "\${RED}Error: Neither curl nor wget found. Please install one of them.\${NC}"
        exit 1
    fi

    chmod +x kli
    echo -e "\${GREEN}✓ Downloaded successfully\${NC}"
}

# Run doctor check
run_doctor() {
    echo ""
    echo -e "\${BLUE}================================\${NC}"
    echo -e "\${BLUE}  Checking AWS Prerequisites\${NC}"
    echo -e "\${BLUE}================================\${NC}"
    echo ""

    if ! ./kli aws doctor; then
        echo ""
        echo -e "\${RED}✗ Prerequisites check failed\${NC}"
        echo -e "\${YELLOW}Please fix the issues above and try again\${NC}"
        rm -f kli
        exit 1
    fi

    echo ""
    echo -e "\${GREEN}✓ Prerequisites check passed\${NC}"
}

# Run installation
run_install() {
    echo ""
    echo -e "\${BLUE}================================\${NC}"
    echo -e "\${BLUE}  Installing Kloudlite on AWS\${NC}"
    echo -e "\${BLUE}================================\${NC}"
    echo ""

    local install_cmd="./kli aws install --installation-key \$INSTALLATION_KEY"

    if [ -n "\$AWS_REGION" ]; then
        install_cmd="\$install_cmd --region \$AWS_REGION"
    fi

    if eval "\$install_cmd"; then
        echo ""
        echo -e "\${GREEN}🎉 Installation complete!\${NC}"
        echo ""
        echo -e "\${BLUE}Installation key: \${INSTALLATION_KEY}\${NC}"
        if [ -n "\$AWS_REGION" ]; then
            echo -e "\${BLUE}Region: \${AWS_REGION}\${NC}"
        fi
    else
        echo ""
        echo -e "\${RED}✗ Installation failed\${NC}"
        rm -f kli
        exit 1
    fi
}

# Cleanup
cleanup() {
    rm -f kli
}

# Main installation flow
main() {
    echo -e "\${BLUE}================================\${NC}"
    echo -e "\${BLUE}  Kloudlite AWS Installer\${NC}"
    echo -e "\${BLUE}================================\${NC}"
    echo ""

    # Detect platform
    platform=$(detect_platform)
    echo -e "\${BLUE}Detected platform: \${platform}\${NC}"
    echo ""

    # Download kli
    download_kli "\$platform"

    # Run doctor check
    run_doctor

    # Run installation
    run_install

    # Cleanup
    cleanup

    echo ""
    echo -e "\${GREEN}Done!\${NC}"
}

main "$@"
`;

export async function GET(_request: NextRequest) {
  return new NextResponse(INSTALL_SCRIPT, {
    headers: {
      'Content-Type': 'text/x-shellscript',
      'Content-Disposition': 'inline; filename="install-aws.sh"',
      'Cache-Control': 'public, max-age=300',
    },
  });
}

export const runtime = 'edge';
