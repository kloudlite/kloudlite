export const runtime = 'nodejs'

import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { getRegistrationSession } from '@/lib/console-auth'
import {
  getUserOrganizations,
  createOrganization,
  isSlugAvailable,
  isValidSlug,
} from '@/lib/console/storage'

/**
 * GET /api/orgs
 * List the current user's organizations
 */
export async function GET() {
  const session = await getRegistrationSession()

  if (!session?.user) {
    return apiError('Unauthorized', 401)
  }

  try {
    const organizations = await getUserOrganizations(session.user.id)
    return NextResponse.json({ organizations })
  } catch (error) {
    return apiCatchError(error, 'Failed to list organizations')
  }
}

/**
 * POST /api/orgs
 * Create a new organization
 * Body: { name: string, slug: string }
 */
export async function POST(request: Request) {
  const session = await getRegistrationSession()

  if (!session?.user) {
    return apiError('Unauthorized', 401)
  }

  try {
    const body = await request.json()
    const { name, slug } = body as { name?: string; slug?: string }

    if (!name || !name.trim()) {
      return apiError('Organization name is required', 400)
    }

    if (!slug || !slug.trim()) {
      return apiError('Organization slug is required', 400)
    }

    if (!isValidSlug(slug)) {
      return apiError(
        'Invalid slug: must be lowercase, alphanumeric, hyphens only, 3-63 chars, start/end with alphanumeric',
        400,
      )
    }

    const available = await isSlugAvailable(slug)
    if (!available) {
      return apiError('Slug is already taken', 409)
    }

    const organization = await createOrganization(session.user.id, name.trim(), slug.trim())
    return NextResponse.json({ organization }, { status: 201 })
  } catch (error) {
    return apiCatchError(error, 'Failed to create organization')
  }
}
