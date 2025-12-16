import { NextRequest, NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationByKey, updateInstallation } from '@/lib/console/supabase-storage-service'

export async function POST(request: NextRequest) {
  try {
    const session = await getRegistrationSession()
    if (!session?.user) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const body = await request.json()
    const { installationKey, cloudProvider, cloudLocation } = body

    if (!installationKey) {
      return NextResponse.json({ error: 'Installation key required' }, { status: 400 })
    }

    if (!cloudProvider || !['aws', 'gcp', 'azure'].includes(cloudProvider)) {
      return NextResponse.json({ error: 'Valid cloud provider required (aws, gcp, azure)' }, { status: 400 })
    }

    if (!cloudLocation) {
      return NextResponse.json({ error: 'Cloud location required' }, { status: 400 })
    }

    // Get installation by key
    const installation = await getInstallationByKey(installationKey)
    if (!installation) {
      return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
    }

    // Verify ownership
    if (installation.userId !== session.user.id) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 403 })
    }

    // Update with cloud provider info
    await updateInstallation(installation.id, {
      cloudProvider,
      cloudLocation,
    })

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('Error setting provider:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
