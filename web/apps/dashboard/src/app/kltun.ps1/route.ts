import { NextRequest, NextResponse } from 'next/server'

/**
 * GET /kltun.ps1
 *
 * Returns an idempotent installation script for kltun on Windows PowerShell
 * Usage: iwr "https://example.com/kltun.ps1?token=TOKEN" -UseBasicParsing | iex
 */
export async function GET(request: NextRequest) {
  // Get token from query params (PowerShell scripts can't easily pass args via piping)
  const token = request.nextUrl.searchParams.get('token') || ''

  // Get the server URL from the request
  const protocol = request.headers.get('x-forwarded-proto') || 'https'
  const host = request.headers.get('host') || request.nextUrl.host
  const serverUrl = `${protocol}://${host}`
  const downloadBaseUrl = `${serverUrl}/api/download/kltun`

  const script = `# kltun Installation and Connection Script for Windows PowerShell
# Run: iwr "${serverUrl}/kltun.ps1?token=YOUR_TOKEN" -UseBasicParsing | iex

$ErrorActionPreference = "Stop"

# Parameters
$Token = "${token}"
$Server = "${serverUrl}"
$DownloadBaseUrl = "${downloadBaseUrl}"

# Validate token
if (-not $Token) {
    Write-Host "Error: Token is required" -ForegroundColor Red
    Write-Host "Usage: iwr \`"${serverUrl}/kltun.ps1?token=YOUR_TOKEN\`" -UseBasicParsing | iex"
    exit 1
}

Write-Host "================================" -ForegroundColor Blue
Write-Host "  kltun Installation" -ForegroundColor Blue
Write-Host "================================" -ForegroundColor Blue
Write-Host ""

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Host "Error: 32-bit systems are not supported" -ForegroundColor Red
    exit 1
}

Write-Host "Platform: windows-$Arch" -ForegroundColor Blue
Write-Host "Server: $Server" -ForegroundColor Blue
Write-Host ""

$DownloadUrl = "$DownloadBaseUrl/windows-$Arch"
$ShaUrl = "$DownloadBaseUrl/windows-$Arch/sha"
$InstallDir = "$env:LOCALAPPDATA\\kltun"
$KltunExe = "$InstallDir\\kltun.exe"

# Check if update is needed
$NeedsUpdate = $true
if (Test-Path $KltunExe) {
    Write-Host "Checking installed kltun version..." -ForegroundColor Blue
    try {
        $RemoteShaResponse = Invoke-RestMethod -Uri $ShaUrl -UseBasicParsing -ErrorAction SilentlyContinue
        $RemoteSha = $RemoteShaResponse.sha256
        if ($RemoteSha) {
            $LocalSha = (Get-FileHash -Path $KltunExe -Algorithm SHA256).Hash.ToLower()
            if ($LocalSha -eq $RemoteSha.ToLower()) {
                Write-Host "kltun is already up-to-date (SHA: $($LocalSha.Substring(0,12))...)" -ForegroundColor Green
                $NeedsUpdate = $false
            } else {
                Write-Host "kltun update available" -ForegroundColor Yellow
                Write-Host "  Local:  $($LocalSha.Substring(0,12))..." -ForegroundColor Blue
                Write-Host "  Remote: $($RemoteSha.Substring(0,12))..." -ForegroundColor Blue
            }
        }
    } catch {
        # Ignore errors checking version, just update
    }
}

if ($NeedsUpdate) {
    Write-Host "Downloading kltun..." -ForegroundColor Blue

    # Create install directory
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # Download to temp file
    $TempFile = [System.IO.Path]::GetTempFileName() + ".exe"
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempFile -UseBasicParsing
    } catch {
        Write-Host "Error: Failed to download kltun: $_" -ForegroundColor Red
        exit 1
    }

    # Stop existing daemon if running
    $kltunProcess = Get-Process -Name "kltun" -ErrorAction SilentlyContinue
    if ($kltunProcess) {
        Write-Host "Stopping existing kltun process..." -ForegroundColor Blue
        Stop-Process -Name "kltun" -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 2
    }

    # Move to install directory
    Write-Host "Installing to $InstallDir..." -ForegroundColor Blue
    Move-Item -Path $TempFile -Destination $KltunExe -Force

    Write-Host "kltun installed successfully!" -ForegroundColor Green

    # Add to PATH if not already present
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        Write-Host "Adding $InstallDir to PATH..." -ForegroundColor Blue
        [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
        $env:Path = "$env:Path;$InstallDir"
    }
}

Write-Host ""

# Start daemon if not running
Write-Host "Checking daemon status..." -ForegroundColor Blue
try {
    $daemonStatus = & $KltunExe daemon status 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Starting kltun daemon..." -ForegroundColor Blue
        Start-Process -FilePath $KltunExe -ArgumentList "daemon", "run" -WindowStyle Hidden
        Start-Sleep -Seconds 3

        # Verify daemon started
        $daemonStatus = & $KltunExe daemon status 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Daemon started" -ForegroundColor Green
        } else {
            Write-Host "Warning: Daemon may not have started properly" -ForegroundColor Yellow
        }
    } else {
        Write-Host "Daemon already running" -ForegroundColor Green
    }
} catch {
    Write-Host "Starting kltun daemon..." -ForegroundColor Blue
    Start-Process -FilePath $KltunExe -ArgumentList "daemon", "run" -WindowStyle Hidden
    Start-Sleep -Seconds 3
}

Write-Host ""
Write-Host "Connecting to VPN..." -ForegroundColor Blue

# Connect to VPN
try {
    & $KltunExe connect --token $Token --server $Server
    if ($LASTEXITCODE -ne 0) {
        throw "Connection failed"
    }
} catch {
    Write-Host "Error: Failed to connect to VPN" -ForegroundColor Red
    Write-Host "You can retry manually with:" -ForegroundColor Yellow
    Write-Host "  $KltunExe daemon run  # Start daemon first"
    Write-Host "  $KltunExe connect --token $Token --server $Server"
    exit 1
}

# Verify connection
Write-Host "Verifying VPN connection..." -ForegroundColor Blue
$Timeout = 60
$Elapsed = 0
$Connected = $false

while ($Elapsed -lt $Timeout) {
    $pingResult = Test-Connection -ComputerName "10.17.0.1" -Count 1 -Quiet -ErrorAction SilentlyContinue
    if ($pingResult) {
        $Connected = $true
        break
    }
    Write-Host "." -NoNewline -ForegroundColor Blue
    Start-Sleep -Seconds 2
    $Elapsed += 2
}
Write-Host ""

if (-not $Connected) {
    Write-Host "Error: VPN connection timeout after $Timeout seconds" -ForegroundColor Red
    Write-Host "Common issues:" -ForegroundColor Yellow
    Write-Host "  - VPN server may be down or unreachable"
    Write-Host "  - Invalid token"
    Write-Host "  - Network connectivity issues"
    Write-Host "  - WireGuard tunnel not fully established"
    exit 1
}

Write-Host ""
Write-Host "Connected successfully!" -ForegroundColor Green
`

  return new NextResponse(script, {
    status: 200,
    headers: {
      'Content-Type': 'text/plain',
      'Content-Disposition': 'inline; filename="install-kltun.ps1"',
      'Cache-Control': 'no-cache',
    },
  })
}
