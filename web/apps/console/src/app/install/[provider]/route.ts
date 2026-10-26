import { NextResponse } from 'next/server'
import { getCloudInstallScript } from '@/lib/cloud-install-script'

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ provider: string }> },
) {
  const { provider } = await params
  const script = getCloudInstallScript(provider, 'install')
  if (!script) return NextResponse.json({ error: 'Unknown provider' }, { status: 400 })
  return new NextResponse(script, { headers: { 'Content-Type': 'text/x-shellscript' } })
}
