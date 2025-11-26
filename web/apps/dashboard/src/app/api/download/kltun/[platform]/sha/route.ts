import { NextRequest, NextResponse } from 'next/server'

const GITHUB_RELEASES_BASE = 'https://github.com/kloudlite/kloudlite/releases'

// Platform to binary name mapping
const PLATFORM_BINARIES: Record<string, string> = {
  'linux-amd64': 'kltun-linux-amd64',
  'linux-arm64': 'kltun-linux-arm64',
  'darwin-amd64': 'kltun-darwin-amd64',
  'darwin-arm64': 'kltun-darwin-arm64',
  'windows-amd64': 'kltun-windows-amd64.exe',
  'windows-arm64': 'kltun-windows-arm64.exe',
}

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ platform: string }> }
) {
  const { platform } = await params
  const searchParams = request.nextUrl.searchParams
  const version = searchParams.get('version') || 'latest'

  // Validate platform
  const binaryName = PLATFORM_BINARIES[platform]
  if (!binaryName) {
    return NextResponse.json(
      {
        error: 'Invalid platform',
        validPlatforms: Object.keys(PLATFORM_BINARIES),
      },
      { status: 400 }
    )
  }

  // Build GitHub release URL for checksum file
  let checksumUrl: string
  let tag: string

  if (version === 'latest') {
    try {
      const releasesResponse = await fetch(
        'https://api.github.com/repos/kloudlite/kloudlite/releases',
        {
          headers: {
            'Accept': 'application/vnd.github+json',
            'User-Agent': 'kloudlite-web',
          },
        }
      )

      if (!releasesResponse.ok) {
        throw new Error('Failed to fetch releases')
      }

      const releases = await releasesResponse.json()

      // Find the latest release with 'kltun-v' tag
      const latestKltunRelease = releases.find(
        (release: { tag_name?: string }) =>
          release.tag_name && release.tag_name.startsWith('kltun-v')
      )

      if (!latestKltunRelease) {
        return NextResponse.json({ error: 'No kltun releases found' }, { status: 404 })
      }

      tag = latestKltunRelease.tag_name
      checksumUrl = `${GITHUB_RELEASES_BASE}/download/${tag}/${binaryName}.sha256`
    } catch (_error) {
      return NextResponse.json(
        { error: 'Failed to determine latest kltun version' },
        { status: 500 }
      )
    }
  } else {
    // Ensure version starts with 'kltun-v'
    tag = version.startsWith('kltun-v') ? version : `kltun-v${version}`
    checksumUrl = `${GITHUB_RELEASES_BASE}/download/${tag}/${binaryName}.sha256`
  }

  // Fetch checksum file for this specific binary
  try {
    const checksumResponse = await fetch(checksumUrl)

    if (!checksumResponse.ok) {
      return NextResponse.json({ error: 'Checksum file not found' }, { status: 404 })
    }

    const checksumContent = await checksumResponse.text()

    // The .sha256 file contains just the SHA256 hash, possibly followed by filename
    // Format could be either:
    // - Just the hash: "abc123..."
    // - Hash with filename: "abc123...  filename"
    const sha256 = checksumContent.trim().split(/\s+/)[0]

    return NextResponse.json(
      {
        platform,
        version: tag,
        binaryName,
        sha256,
      },
      {
        status: 200,
        headers: {
          'Cache-Control': 'public, max-age=300', // Cache for 5 minutes
        },
      }
    )
  } catch (_error) {
    return NextResponse.json(
      { error: 'Failed to fetch or parse checksums' },
      { status: 500 }
    )
  }
}

export const runtime = 'edge'
