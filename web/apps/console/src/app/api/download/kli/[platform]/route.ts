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

  // Validate platform
  if (!VALID_PLATFORMS.includes(platform)) {
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
  const isWindows = platform.startsWith('windows');
  const assetName = isWindows ? `kli-${platform}.exe` : `kli-${platform}`;

  const asset = release.assets.find((a) => a.name === assetName);

  if (!asset) {
    return NextResponse.json(
      {
        error: `Binary not found for platform: ${platform}`,
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
