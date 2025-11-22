// Username utility functions for client-side use

/**
 * Generate a valid Kubernetes username from an email address
 * @param email - Email address to convert
 * @returns Valid k8s username (3-63 chars, lowercase alphanumeric with hyphens)
 */
export function generateUsernameFromEmail(email: string): string {
  if (!email) return ''

  // Extract part before @
  const parts = email.split('@')
  if (parts.length === 0) return ''

  let username = parts[0]

  // Replace dots and underscores with hyphens
  username = username.replace(/[._]/g, '-')

  // Convert to lowercase
  username = username.toLowerCase()

  // Remove any invalid characters (keep only lowercase alphanumeric and hyphens)
  username = username.replace(/[^a-z0-9-]/g, '')

  // Ensure it starts and ends with alphanumeric
  username = username.replace(/^-+|-+$/g, '')

  // Ensure minimum length
  if (username.length < 3) {
    username = username + '-user'
  }

  // Ensure maximum length
  if (username.length > 63) {
    username = username.substring(0, 63)
  }

  // Trim trailing hyphens again in case we truncated
  username = username.replace(/-+$/, '')

  return username
}
