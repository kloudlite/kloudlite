import { NextRequest, NextResponse } from 'next/server';

const GITHUB_RELEASES_BASE = 'https://github.com/kloudlite/kloudlite/releases';

// Platform to binary name mapping
const PLATFORM_BINARIES: Record<string, string> = {
  'linux-amd64': 'kltun-linux-amd64',
  'linux-arm64': 'kltun-linux-arm64',
  'darwin-amd64': 'kltun-darwin-amd64',
  'darwin-arm64': 'kltun-darwin-arm64',
  'windows-amd64': 'kltun-windows-amd64.exe',
  'windows-arm64': 'kltun-windows-arm64.exe',
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
    // For 'latest', we need to find the latest kltun-specific release
    // GitHub's /latest endpoint returns the latest release overall, not filtered by tag prefix
    // So we fetch all releases and find the latest kltun release
    try {
      const releasesResponse = await fetch(
        'https://api.github.com/repos/kloudlite/kloudlite/releases',
        {
          headers: {
            'Accept': 'application/vnd.github+json',
            'User-Agent': 'kloudlite-web',
          },
        }
      );

      if (!releasesResponse.ok) {
        throw new Error('Failed to fetch releases');
      }

      const releases = await releasesResponse.json();

      // Find the latest release with 'kltun-v' tag
      const latestKltunRelease = releases.find((release: any) =>
        release.tag_name && release.tag_name.startsWith('kltun-v')
      );

      if (!latestKltunRelease) {
        return NextResponse.json(
          { error: 'No kltun releases found' },
          { status: 404 }
        );
      }

      downloadUrl = `${GITHUB_RELEASES_BASE}/download/${latestKltunRelease.tag_name}/${binaryName}`;
    } catch (error) {
      return NextResponse.json(
        { error: 'Failed to determine latest kltun version' },
        { status: 500 }
      );
    }
  } else {
    // Ensure version starts with 'kltun-v'
    const tag = version.startsWith('kltun-v') ? version : `kltun-v${version}`;
    downloadUrl = `${GITHUB_RELEASES_BASE}/download/${tag}/${binaryName}`;
  }

  // Redirect to GitHub releases
  return NextResponse.redirect(downloadUrl, 302);
}

export const runtime = 'edge';
