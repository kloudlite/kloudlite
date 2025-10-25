import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/registration-auth'
import { getInstallationById, deleteInstallation } from '@/lib/registration/supabase-storage-service'

/**
 * Delete installation API route
 */
export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  // Fetch the installation
  const installation = await getInstallationById(id)

  if (!installation) {
    return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
  }

  // Verify user owns this installation
  if (installation.userId !== session.user.id) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
  }

  // Delete the installation
  try {
    await deleteInstallation(id)
    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('Error deleting installation:', error)
    return NextResponse.json(
      { error: 'Failed to delete installation' },
      { status: 500 }
    )
  }
}
