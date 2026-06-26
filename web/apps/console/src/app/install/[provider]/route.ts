import { NextRequest, NextResponse } from 'next/server'

const SCRIPTS: Record<string, string> = {
  'aws': `#!/bin/bash
set -e
KEY=""; REGION=""; LOC=""; ZONE=""
while [ $# -gt 0 ]; do case $1 in --key) KEY="$2"; shift 2;; --region) REGION="$2"; shift 2;; --location) LOC="$2"; shift 2;; --zone) ZONE="$2"; shift 2;; *) echo "Unknown: $1"; exit 1;; esac; done
[ -z "$KEY" ] && { echo "--key required"; exit 1; }
echo -e "\\033[0;32mDownloading kli...\\033[0m"
curl -fsSL "https://get.khost.dev/api/download/kli/linux-amd64" -o /usr/local/bin/kli && chmod +x /usr/local/bin/kli
echo -e "\\033[0;32mInstalling...\\033[0m"
kli install aws \${REGION:+--region "$REGION"} --key "$KEY"
echo -e "\\033[0;32mDone\\033[0m"`,
  'azure': `#!/bin/bash
set -e
KEY=""; REGION=""; LOC=""; ZONE=""
while [ $# -gt 0 ]; do case $1 in --key) KEY="$2"; shift 2;; --region) REGION="$2"; shift 2;; --location) LOC="$2"; shift 2;; --zone) ZONE="$2"; shift 2;; *) echo "Unknown: $1"; exit 1;; esac; done
[ -z "$KEY" ] && { echo "--key required"; exit 1; }
echo -e "\\033[0;32mDownloading kli...\\033[0m"
curl -fsSL "https://get.khost.dev/api/download/kli/linux-amd64" -o /usr/local/bin/kli && chmod +x /usr/local/bin/kli
echo -e "\\033[0;32mInstalling...\\033[0m"
kli install azure \${LOC:+--location "$LOC"} --key "$KEY"
echo -e "\\033[0;32mDone\\033[0m"`,
  'gcp': `#!/bin/bash
set -e
KEY=""; REGION=""; LOC=""; ZONE=""
while [ $# -gt 0 ]; do case $1 in --key) KEY="$2"; shift 2;; --region) REGION="$2"; shift 2;; --location) LOC="$2"; shift 2;; --zone) ZONE="$2"; shift 2;; *) echo "Unknown: $1"; exit 1;; esac; done
[ -z "$KEY" ] && { echo "--key required"; exit 1; }
echo -e "\\033[0;32mDownloading kli...\\033[0m"
curl -fsSL "https://get.khost.dev/api/download/kli/linux-amd64" -o /usr/local/bin/kli && chmod +x /usr/local/bin/kli
echo -e "\\033[0;32mInstalling...\\033[0m"
kli install gcp \${REGION:+--region "$REGION"} \${ZONE:+--zone "$ZONE"} --key "$KEY"
echo -e "\\033[0;32mDone\\033[0m"`,
}

export async function GET(_req: NextRequest, { params }: { params: Promise<{ provider: string }> }) {
  const { provider } = await params
  const script = SCRIPTS[provider]
  if (!script) return NextResponse.json({ error: 'Unknown provider' }, { status: 400 })
  return new NextResponse(script, { headers: { 'Content-Type': 'text/x-shellscript' } })
}
