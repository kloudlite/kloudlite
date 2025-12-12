import { NextRequest, NextResponse } from 'next/server';

const UNINSTALL_SCRIPT = `#!/bin/bash
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
    echo "Usage: curl -fsSL https://get.khost.dev/uninstall/aws | bash -s -- --key YOUR_KEY [--region REGION]"
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

# Check if kli binary is up to date using MD5 checksum
check_kli_version() {
    local platform=\$1
    local md5_url="\${DOWNLOAD_BASE_URL}/\${platform}.md5"
    local expected_md5=""

    # Download expected MD5 checksum
    if command -v curl &> /dev/null; then
        expected_md5=\$(curl -fsSL "\${md5_url}" 2>/dev/null | awk '{print \$1}')
    elif command -v wget &> /dev/null; then
        expected_md5=\$(wget -qO- "\${md5_url}" 2>/dev/null | awk '{print \$1}')
    fi

    if [ -z "\$expected_md5" ]; then
        echo "unknown"
        return
    fi

    # Check if kli exists in current directory or PATH
    local kli_path=""
    if [ -f "./kli" ]; then
        kli_path="./kli"
    elif command -v kli &> /dev/null; then
        kli_path=\$(command -v kli)
    fi

    if [ -z "\$kli_path" ]; then
        echo "not_found"
        return
    fi

    # Calculate actual MD5
    local actual_md5=""
    if command -v md5sum &> /dev/null; then
        actual_md5=\$(md5sum "\$kli_path" | awk '{print \$1}')
    elif command -v md5 &> /dev/null; then
        actual_md5=\$(md5 -q "\$kli_path")
    else
        echo "unknown"
        return
    fi

    if [ "\$expected_md5" = "\$actual_md5" ]; then
        echo "up_to_date:\$kli_path"
    else
        echo "outdated:\$kli_path"
    fi
}

# Download kli binary
download_kli() {
    local platform=\$1
    local download_url="\${DOWNLOAD_BASE_URL}/\${platform}"

    # Check if we already have an up-to-date version
    local version_status=\$(check_kli_version "\$platform")
    local status=\$(echo "\$version_status" | cut -d: -f1)
    local kli_path=\$(echo "\$version_status" | cut -d: -f2)

    case "\$status" in
        "up_to_date")
            echo -e "\${GREEN}✓ kli is already up to date at \${kli_path}\${NC}"
            if [ "\$kli_path" != "./kli" ]; then
                cp "\$kli_path" ./kli
                chmod +x ./kli
            fi
            return
            ;;
        "outdated")
            echo -e "\${YELLOW}Existing kli at \${kli_path} is outdated, downloading latest...\${NC}"
            ;;
        *)
            echo -e "\${BLUE}Downloading kli for \${platform}...\${NC}"
            ;;
    esac

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

# Run uninstallation
run_uninstall() {
    echo ""
    echo -e "\${BLUE}================================\${NC}"
    echo -e "\${BLUE}  Uninstalling Kloudlite from AWS\${NC}"
    echo -e "\${BLUE}================================\${NC}"
    echo ""

    local uninstall_cmd="./kli aws uninstall --installation-key \$INSTALLATION_KEY"

    if [ -n "\$AWS_REGION" ]; then
        uninstall_cmd="\$uninstall_cmd --region \$AWS_REGION"
    fi

    echo -e "\${YELLOW}This will delete all resources with installation key: \${INSTALLATION_KEY}\${NC}"
    if [ -n "\$AWS_REGION" ]; then
        echo -e "\${YELLOW}Region: \${AWS_REGION}\${NC}"
    fi
    echo ""

    if eval "\$uninstall_cmd"; then
        echo ""
        echo -e "\${GREEN}✓ Uninstallation complete!\${NC}"
        echo ""
        echo -e "\${BLUE}All resources with installation key '\${INSTALLATION_KEY}' have been removed\${NC}"
    else
        echo ""
        echo -e "\${RED}✗ Uninstallation failed\${NC}"
        rm -f kli
        exit 1
    fi
}

# Cleanup
cleanup() {
    rm -f kli
}

# Main uninstallation flow
main() {
    echo -e "\${BLUE}================================\${NC}"
    echo -e "\${BLUE}  Kloudlite AWS Uninstaller\${NC}"
    echo -e "\${BLUE}================================\${NC}"
    echo ""

    # Detect platform
    platform=$(detect_platform)
    echo -e "\${BLUE}Detected platform: \${platform}\${NC}"
    echo ""

    # Download kli
    download_kli "\$platform"

    # Run uninstallation
    run_uninstall

    # Cleanup
    cleanup

    echo ""
    echo -e "\${GREEN}Done!\${NC}"
}

main "$@"
`;

export async function GET(_request: NextRequest) {
  return new NextResponse(UNINSTALL_SCRIPT, {
    headers: {
      'Content-Type': 'text/x-shellscript',
      'Content-Disposition': 'inline; filename="uninstall-aws.sh"',
      'Cache-Control': 'public, max-age=300',
    },
  });
}

export const runtime = 'edge';
