import { cache } from 'react'
import { requireInstallationAccess } from '@/lib/console/authorization'
import { getInstallationById } from '@/lib/console/storage'

export const cachedInstallationAccess = cache(async (installationId: string) => {
  return requireInstallationAccess(installationId)
})

export const cachedInstallationById = cache(async (installationId: string) => {
  return getInstallationById(installationId)
})
