import { NextRequest, NextResponse } from 'next/server';

const GITHUB_REPO = 'kloudlite/kloudlite';
const VALID_PLATFORMS = [
  'linux-amd64',
  'linux-arm64',
  'darwin-amd64',
  'darwin-arm64',
  'windows-amd64',
  'windows-arm64',
];

interface GitHubRelease {
  tag_name: string;
  assets: Array<{
    name: string;
    browser_download_url: string;
  }>;
}

async function getLatestKliRelease(): Promise<GitHubRelease | null> {
  try {
    // Fetch releases and find the latest kli release
    const response = await fetch(
      `https://api.github.com/repos/${GITHUB_REPO}/releases?per_page=50`,
      {
        headers: {
          Accept: 'application/vnd.github.v3+json',
          ...(process.env.GITHUB_TOKEN && {
            Authorization: `token ${process.env.GITHUB_TOKEN}`,
          }),
        },
        next: { revalidate: 300 }, // Cache for 5 minutes
      }
    );

    if (!response.ok) {
      console.error('Failed to fetch releases:', response.status);
      return null;
    }

    const releases: GitHubRelease[] = await response.json();

    // Find the latest kli release (tag starts with kli-v)
    const kliRelease = releases.find((r) => r.tag_name.startsWith('kli-v'));
    return kliRelease || null;
  } catch (error) {
    console.error('Error fetching releases:', error);
    return null;
  }
}

async function getKliReleaseByVersion(
  version: string
): Promise<GitHubRelease | null> {
  try {
    const tag = version.startsWith('kli-v') ? version : `kli-v${version}`;
    const response = await fetch(
      `https://api.github.com/repos/${GITHUB_REPO}/releases/tags/${tag}`,
      {
        headers: {
          Accept: 'application/vnd.github.v3+json',
          ...(process.env.GITHUB_TOKEN && {
            Authorization: `token ${process.env.GITHUB_TOKEN}`,
          }),
        },
        next: { revalidate: 300 },
      }
    );

    if (!response.ok) {
      return null;
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching release by version:', error);
    return null;
  }
}

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ platform: string }> }
) {
  const { platform } = await params;

  // Check if this is a request for MD5 checksum
  const isMd5Request = platform.endsWith('.md5');
  const actualPlatform = isMd5Request ? platform.replace('.md5', '') : platform;

  // Validate platform
  if (!VALID_PLATFORMS.includes(actualPlatform)) {
    return NextResponse.json(
      {
        error: 'Invalid platform',
        valid_platforms: VALID_PLATFORMS,
      },
      { status: 400 }
    );
  }

  // Check for version query parameter
  const { searchParams } = new URL(request.url);
  const version = searchParams.get('version');

  // Get the release
  const release = version
    ? await getKliReleaseByVersion(version)
    : await getLatestKliRelease();

  if (!release) {
    return NextResponse.json(
      { error: 'No kli release found' },
      { status: 404 }
    );
  }

  // Find the asset for the requested platform
  // Asset naming convention: kli-{os}-{arch} or kli-{os}-{arch}.exe for windows
  const isWindows = actualPlatform.startsWith('windows');
  const baseName = isWindows ? `kli-${actualPlatform}.exe` : `kli-${actualPlatform}`;
  const assetName = isMd5Request ? `${baseName}.md5` : baseName;

  const asset = release.assets.find((a) => a.name === assetName);

  if (!asset) {
    // If MD5 file not found, return 404 without error details
    if (isMd5Request) {
      return NextResponse.json(
        { error: `MD5 checksum not found for platform: ${actualPlatform}` },
        { status: 404 }
      );
    }
    return NextResponse.json(
      {
        error: `Binary not found for platform: ${actualPlatform}`,
        available_assets: release.assets.map((a) => a.name),
        release_tag: release.tag_name,
      },
      { status: 404 }
    );
  }

  // Redirect to the GitHub download URL
  return NextResponse.redirect(asset.browser_download_url, {
    status: 302,
    headers: {
      'Cache-Control': 'public, max-age=300',
    },
  });
}
