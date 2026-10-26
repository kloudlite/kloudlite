type Operation = 'install' | 'uninstall'

const PROVIDER_FLAGS: Record<string, string> = {
  aws: '${REGION:+--region "$REGION"}',
  azure: '${LOC:+--location "$LOC"}',
  gcp: '${REGION:+--region "$REGION"} ${ZONE:+--zone "$ZONE"}',
}

export function getCloudInstallScript(provider: string, operation: Operation): string | null {
  const flags = PROVIDER_FLAGS[provider]
  if (!flags) return null

  const verb = operation === 'install' ? 'Installing' : 'Uninstalling'
  return `#!/bin/bash
set -e
KEY=""; REGION=""; LOC=""; ZONE=""
while [ $# -gt 0 ]; do case $1 in --key) KEY="$2"; shift 2;; --region) REGION="$2"; shift 2;; --location) LOC="$2"; shift 2;; --zone) ZONE="$2"; shift 2;; *) echo "Unknown: $1"; exit 1;; esac; done
[ -z "$KEY" ] && { echo "--key required"; exit 1; }
echo -e "\\033[0;32mDownloading kli...\\033[0m"
curl -fsSL "https://get.khost.dev/api/download/kli/linux-amd64" -o /usr/local/bin/kli && chmod +x /usr/local/bin/kli
echo -e "\\033[0;32m${verb}...\\033[0m"
kli ${operation} ${provider} ${flags} --key "$KEY"
echo -e "\\033[0;32mDone\\033[0m"`
}
