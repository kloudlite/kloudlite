/**
 * Format error messages to be more user-friendly
 * Converts technical errors to readable messages
 */
export function formatErrorMessage(error: string): string {
  // Handle common gRPC/technical errors
  const errorMap: Record<string, string> = {
    'user already exists': 'An account with this email already exists',
    'already exists': 'This email is already registered',
    'not valid credentials': 'Invalid email or password',
    'invalid credentials': 'Invalid email or password',
    'failed to create user': 'Unable to create account. Please try again',
    'failed to send': 'Unable to send email. Please try again',
    'connection refused': 'Unable to connect to server. Please try again',
    'network error': 'Network error. Please check your connection',
    'timeout': 'Request timed out. Please try again',
    'internal server error': 'Something went wrong. Please try again',
    'unauthorized': 'You are not authorized to perform this action',
    'forbidden': 'Access denied',
  }

  // Check for error patterns and return friendly message
  const lowerError = error.toLowerCase()
  for (const [key, value] of Object.entries(errorMap)) {
    if (lowerError.includes(key)) {
      return value
    }
  }

  // Format the error message - capitalize first letter and ensure proper punctuation
  let formatted = error.trim()
  if (formatted.length > 0) {
    formatted = formatted.charAt(0).toUpperCase() + formatted.slice(1)
    if (!formatted.endsWith('.') && !formatted.endsWith('!') && !formatted.endsWith('?')) {
      formatted += '.'
    }
  }

  return formatted || 'An unexpected error occurred. Please try again.'
}