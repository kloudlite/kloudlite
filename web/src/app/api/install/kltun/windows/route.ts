import { NextRequest, NextResponse } from 'next/server';

const INSTALL_SCRIPT = `# Kloudlite Tunnel (kltun) - Windows Installation Script
$ErrorActionPreference = "Stop"

# Base URL for downloads - will be replaced with actual subdomain
$DOWNLOAD_BASE_URL = if ($env:KLTUN_SERVER_URL) { $env:KLTUN_SERVER_URL } else { "https://KLTUN_BASE_URL/api/download/kltun" }

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

# Download kltun binary
function Download-Kltun {
    param(
        [string]$Platform,
        [string]$Version = "latest"
    )

    $downloadUrl = "$DOWNLOAD_BASE_URL/$Platform"
    if ($Version -ne "latest") {
        $downloadUrl += "?version=$Version"
    }

    Write-ColorOutput Blue "Downloading kltun for $Platform..."

    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile "kltun.exe" -UseBasicParsing
        Write-ColorOutput Green "✓ Downloaded successfully"
    } catch {
        Write-ColorOutput Red "Error: Failed to download kltun"
        Write-ColorOutput Red $_.Exception.Message
        exit 1
    }
}

# Install kltun to system
function Install-Kltun {
    $installDir = "$env:LOCALAPPDATA\\kloudlite\\bin"

    # Create directory if it doesn't exist
    if (!(Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    # Move binary
    Move-Item -Path "kltun.exe" -Destination "$installDir\\kltun.exe" -Force
    Write-ColorOutput Green "✓ Installed to $installDir\\kltun.exe"

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
        $version = & kltun version 2>&1
        if ($LASTEXITCODE -ne 0) {
            $version = "unknown"
        }
        Write-ColorOutput Green "✓ kltun installed successfully!"
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

Kloudlite Tunnel (kltun) - Windows Installation Script

Usage:
  iwr -useb https://KLTUN_BASE_URL/kltun/windows | iex
  iwr -useb https://KLTUN_BASE_URL/kltun/windows | iex -Version 0.1.0

Examples:
  # Install latest version
  iwr -useb https://KLTUN_BASE_URL/kltun/windows | iex

  # Install specific version
  \`$env:KLTUN_VERSION="0.1.0"; iwr -useb https://KLTUN_BASE_URL/kltun/windows | iex

After installation:
  kltun --help                Show kltun help
  kltun daemon                Start daemon
  kltun connect               Connect to VPN
  kltun install-ca            Install CA certificates
  kltun hosts add             Manage hosts file entries

"@
}

# Main installation flow
function Main {
    param([string]$Version = "latest")

    # Check for version in environment variable
    if ($env:KLTUN_VERSION) {
        $Version = $env:KLTUN_VERSION
    }

    Write-ColorOutput Blue "================================"
    Write-ColorOutput Blue "  Kloudlite Tunnel (kltun)"
    Write-ColorOutput Blue "================================"
    Write-Output ""

    # Detect platform
    $arch = Get-Architecture
    $platform = "windows-$arch"
    Write-ColorOutput Blue "Detected platform: $platform"
    Write-Output ""

    # Download
    Download-Kltun -Platform $platform -Version $Version
    Write-Output ""

    # Install
    Install-Kltun
    Write-Output ""

    # Verify
    if (Test-Installation) {
        Write-Output ""
        Write-ColorOutput Green "🎉 Installation complete!"
        Write-Output ""
        Write-ColorOutput Blue "Quick start:"
        Write-Output "  kltun --help          # Show help"
        Write-Output "  kltun daemon          # Start daemon"
        Write-Output "  kltun connect         # Connect to VPN"
        Write-Output ""
        Write-ColorOutput Blue "Documentation:"
        Write-Output "  https://github.com/kloudlite/kloudlite/tree/development/api/cmd/kltun"
    } else {
        exit 1
    }
}

# Run main function
Main
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
      'Content-Type': 'text/plain',
      'Content-Disposition': 'inline; filename="install.ps1"',
      'Cache-Control': 'public, max-age=300', // Cache for 5 minutes
    },
  });
}

export const runtime = 'edge';
