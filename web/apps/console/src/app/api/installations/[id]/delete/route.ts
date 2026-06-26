import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireInstallationOwner } from '@/lib/console/authorization'
import {
  getInstallationById,
  deleteInstallation,
} from '@/lib/console/storage'

/**
 * Delete installation API route
 */
export async function DELETE(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  try {
    await requireInstallationOwner(id)

    const installation = await getInstallationById(id)
    if (!installation) {
      return apiError('Installation not found', 404)
    }

    if (
      installation.deployJobOperation === 'uninstall' &&
      (installation.deployJobStatus === 'running' || installation.deployJobStatus === 'pending')
    ) {
      return apiError('Cannot delete while uninstall is in progress', 409)
    }

    await deleteInstallation(id)
    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('Error deleting installation:', error)
    return apiCatchError(error, 'Failed to delete installation')
  }
}
