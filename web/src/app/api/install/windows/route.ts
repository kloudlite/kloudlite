import { NextRequest, NextResponse } from 'next/server';

const INSTALL_SCRIPT = `# Kloudlite Installer (kli) - Windows Installation Script
$ErrorActionPreference = "Stop"

# Base URL for downloads
$DOWNLOAD_BASE_URL = "https://console.kloudlite.io/api/download/kli"

# Colors
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

# Detect Architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Write-ColorOutput Red "Error: Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Download kli binary
function Download-Kli {
    param(
        [string]$Platform,
        [string]$Version = "latest"
    )

    $downloadUrl = "$DOWNLOAD_BASE_URL/$Platform"
    if ($Version -ne "latest") {
        $downloadUrl += "?version=$Version"
    }

    Write-ColorOutput Blue "Downloading kli for $Platform..."

    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile "kli.exe" -UseBasicParsing
        Write-ColorOutput Green "✓ Downloaded successfully"
    } catch {
        Write-ColorOutput Red "Error: Failed to download kli"
        Write-ColorOutput Red $_.Exception.Message
        exit 1
    }
}

# Install kli to system
function Install-Kli {
    $installDir = "$env:LOCALAPPDATA\\kloudlite\\bin"

    # Create directory if it doesn't exist
    if (!(Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    # Move binary
    Move-Item -Path "kli.exe" -Destination "$installDir\\kli.exe" -Force
    Write-ColorOutput Green "✓ Installed to $installDir\\kli.exe"

    # Add to PATH if not already there
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$installDir*") {
        Write-ColorOutput Yellow "Adding $installDir to PATH..."
        [Environment]::SetEnvironmentVariable(
            "Path",
            "$userPath;$installDir",
            "User"
        )
        $env:Path += ";$installDir"
        Write-ColorOutput Green "✓ Added to PATH"
        Write-ColorOutput Yellow "Note: You may need to restart your terminal for PATH changes to take effect"
    }
}

# Verify installation
function Test-Installation {
    try {
        $version = & kli version 2>&1
        Write-ColorOutput Green "✓ kli installed successfully!"
        Write-ColorOutput Blue "Version: $version"
        return $true
    } catch {
        Write-ColorOutput Red "✗ Installation failed"
        return $false
    }
}

# Print usage
function Show-Usage {
    Write-Output @"

Kloudlite Installer (kli) - Windows Installation Script

Usage:
  iwr -useb https://console.kloudlite.io/api/install/windows | iex
  iwr -useb https://console.kloudlite.io/api/install/windows | iex -Version 0.1.0

Examples:
  # Install latest version
  iwr -useb https://console.kloudlite.io/api/install/windows | iex

  # Install specific version
  `$env:KLI_VERSION="0.1.0"; iwr -useb https://console.kloudlite.io/api/install/windows | iex

After installation:
  kli --help          Show kli help
  kli aws doctor      Check AWS prerequisites
  kli aws install     Install Kloudlite on AWS

"@
}

# Main installation flow
function Main {
    param([string]$Version = "latest")

    # Check for version in environment variable
    if ($env:KLI_VERSION) {
        $Version = $env:KLI_VERSION
    }

    Write-ColorOutput Blue "================================"
    Write-ColorOutput Blue "  Kloudlite Installer (kli)"
    Write-ColorOutput Blue "================================"
    Write-Output ""

    # Detect platform
    $arch = Get-Architecture
    $platform = "windows-$arch"
    Write-ColorOutput Blue "Detected platform: $platform"
    Write-Output ""

    # Download
    Download-Kli -Platform $platform -Version $Version
    Write-Output ""

    # Install
    Install-Kli
    Write-Output ""

    # Verify
    if (Test-Installation) {
        Write-Output ""
        Write-ColorOutput Green "🎉 Installation complete!"
        Write-Output ""
        Write-ColorOutput Blue "Quick start:"
        Write-Output "  kli aws doctor        # Check AWS prerequisites"
        Write-Output "  kli aws install       # Install Kloudlite on AWS"
        Write-Output ""
        Write-ColorOutput Blue "Documentation:"
        Write-Output "  https://github.com/kloudlite/kloudlite/tree/development/api/cmd/kli"
    } else {
        exit 1
    }
}

# Run main function
Main
`;

export async function GET(request: NextRequest) {
  return new NextResponse(INSTALL_SCRIPT, {
    headers: {
      'Content-Type': 'text/plain',
      'Content-Disposition': 'inline; filename="install.ps1"',
      'Cache-Control': 'public, max-age=300', // Cache for 5 minutes
    },
  });
}

export const runtime = 'edge';
