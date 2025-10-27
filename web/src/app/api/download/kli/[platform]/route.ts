import { NextRequest, NextResponse } from 'next/server';

const GITHUB_RELEASES_BASE = 'https://github.com/kloudlite/kloudlite/releases';

// Platform to binary name mapping
const PLATFORM_BINARIES: Record<string, string> = {
  'linux-amd64': 'kli-linux-amd64',
  'linux-arm64': 'kli-linux-arm64',
  'darwin-amd64': 'kli-darwin-amd64',
  'darwin-arm64': 'kli-darwin-arm64',
  'windows-amd64': 'kli-windows-amd64.exe',
  'windows-arm64': 'kli-windows-arm64.exe',
};

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ platform: string }> }
) {
  const { platform } = await params;
  const searchParams = request.nextUrl.searchParams;
  const version = searchParams.get('version') || 'latest';

  // Validate platform
  const binaryName = PLATFORM_BINARIES[platform];
  if (!binaryName) {
    return NextResponse.json(
      {
        error: 'Invalid platform',
        validPlatforms: Object.keys(PLATFORM_BINARIES),
      },
      { status: 400 }
    );
  }

  // Build GitHub release URL
  let downloadUrl: string;
  if (version === 'latest') {
    downloadUrl = `${GITHUB_RELEASES_BASE}/latest/download/${binaryName}`;
  } else {
    // Ensure version starts with 'kli-v'
    const tag = version.startsWith('kli-v') ? version : `kli-v${version}`;
    downloadUrl = `${GITHUB_RELEASES_BASE}/download/${tag}/${binaryName}`;
  }

  // Redirect to GitHub releases
  return NextResponse.redirect(downloadUrl, 302);
}

export const runtime = 'edge';
