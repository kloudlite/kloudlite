import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Formats a resource name by extracting username and resource name from the format {username}--{resourcename}
 * Returns formatted as {username}/{resourcename}
 * If the name doesn't contain '--', returns the name as-is
 */
export function formatResourceName(fullName: string): string {
  const parts = fullName.split('--')
  if (parts.length === 2) {
    return `${parts[0]}/${parts[1]}`
  }
  return fullName
}

/**
 * Extracts just the resource name part from {username}--{resourcename}
 * If the name doesn't contain '--', returns the name as-is
 */
export function getResourceName(fullName: string): string {
  const parts = fullName.split('--')
  if (parts.length === 2) {
    return parts[1]
  }
  return fullName
}

/**
 * Extracts the username part from {username}--{resourcename}
 * If the name doesn't contain '--', returns null
 */
export function getResourceOwner(fullName: string): string | null {
  const parts = fullName.split('--')
  if (parts.length === 2) {
    return parts[0]
  }
  return null
}
