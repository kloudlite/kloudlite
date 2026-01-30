/**
 * In-memory rate limiter for magic link requests
 * Prevents abuse by limiting requests per email and IP address
 */

interface RateLimitEntry {
  count: number
  resetAt: number
}

// In-memory storage for rate limiting
const emailRateLimits = new Map<string, RateLimitEntry>()
const ipRateLimits = new Map<string, RateLimitEntry>()

// Configuration
const isDevelopment = process.env.NODE_ENV === 'development'
const EMAIL_LIMIT = isDevelopment ? 100 : 3 // Max requests per email (100 in dev, 3 in prod)
const EMAIL_WINDOW = 15 * 60 * 1000 // 15 minutes in milliseconds
const IP_LIMIT = isDevelopment ? 1000 : 10 // Max requests per IP (1000 in dev, 10 in prod)
const IP_WINDOW = 60 * 60 * 1000 // 1 hour in milliseconds

/**
 * Check if an email has exceeded rate limit
 * @param email - Email address to check
 * @returns true if email is within rate limit, false if exceeded
 */
export function checkEmailRateLimit(email: string): boolean {
  const now = Date.now()
  const entry = emailRateLimits.get(email)

  if (!entry) {
    return true // No entry, allow request
  }

  if (now > entry.resetAt) {
    // Window expired, reset
    emailRateLimits.delete(email)
    return true
  }

  return entry.count < EMAIL_LIMIT
}

/**
 * Check if an IP address has exceeded rate limit
 * @param ip - IP address to check
 * @returns true if IP is within rate limit, false if exceeded
 */
export function checkIPRateLimit(ip: string): boolean {
  const now = Date.now()
  const entry = ipRateLimits.get(ip)

  if (!entry) {
    return true // No entry, allow request
  }

  if (now > entry.resetAt) {
    // Window expired, reset
    ipRateLimits.delete(ip)
    return true
  }

  return entry.count < IP_LIMIT
}

/**
 * Record a request for an email address
 * @param email - Email address to record
 */
export function recordEmailRequest(email: string): void {
  const now = Date.now()
  const entry = emailRateLimits.get(email)

  if (!entry || now > entry.resetAt) {
    // Create new entry or reset expired one
    emailRateLimits.set(email, {
      count: 1,
      resetAt: now + EMAIL_WINDOW,
    })
  } else {
    // Increment existing entry
    entry.count++
  }
}

/**
 * Record a request for an IP address
 * @param ip - IP address to record
 */
export function recordIPRequest(ip: string): void {
  const now = Date.now()
  const entry = ipRateLimits.get(ip)

  if (!entry || now > entry.resetAt) {
    // Create new entry or reset expired one
    ipRateLimits.set(ip, {
      count: 1,
      resetAt: now + IP_WINDOW,
    })
  } else {
    // Increment existing entry
    entry.count++
  }
}

/**
 * Clean up expired entries from rate limit maps
 * Should be called periodically to prevent memory leaks
 */
export function cleanupExpiredEntries(): void {
  const now = Date.now()

  // Clean email rate limits
  for (const [email, entry] of emailRateLimits.entries()) {
    if (now > entry.resetAt) {
      emailRateLimits.delete(email)
    }
  }

  // Clean IP rate limits
  for (const [ip, entry] of ipRateLimits.entries()) {
    if (now > entry.resetAt) {
      ipRateLimits.delete(ip)
    }
  }
}

// Run cleanup every 5 minutes
if (typeof setInterval !== 'undefined') {
  setInterval(cleanupExpiredEntries, 5 * 60 * 1000)
}
