export const runtime = 'nodejs'

import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireOrgAccess, requireOrgOwner } from '@/lib/console/authorization'
import {
  getOrganizationById,
  updateOrganization,
  deleteOrganization,
} from '@/lib/console/storage'
import { cancelSubscriptionForOrg } from '@/lib/console/storage'

/**
 * GET /api/orgs/[orgId]
 * Get organization details
 */
export async function GET(
  _request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    await requireOrgAccess(orgId)
    const organization = await getOrganizationById(orgId)

    if (!organization) {
      return apiError('Organization not found', 404)
    }

    return NextResponse.json({ organization })
  } catch (error) {
    return apiCatchError(error, 'Failed to get organization')
  }
}

/**
 * PATCH /api/orgs/[orgId]
 * Update organization name (owner only)
 * Body: { name: string }
 */
export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    await requireOrgOwner(orgId)

    const body = await request.json()
    const { name } = body as { name?: string }

    if (!name || !name.trim()) {
      return apiError('Organization name is required', 400)
    }

    const organization = await updateOrganization(orgId, { name: name.trim() })
    return NextResponse.json({ organization })
  } catch (error) {
    return apiCatchError(error, 'Failed to update organization')
  }
}

/**
 * DELETE /api/orgs/[orgId]
 * Delete organization (owner only)
 */
export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    await requireOrgOwner(orgId)

    // Cancel any active subscription before deleting
    await cancelSubscriptionForOrg(orgId)
    await deleteOrganization(orgId)

    return NextResponse.json({ success: true })
  } catch (error) {
    return apiCatchError(error, 'Failed to delete organization')
  }
}
