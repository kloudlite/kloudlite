/**
 * API endpoint to request a magic link
 * POST /api/auth/magic-link/send
 */

import { NextRequest, NextResponse } from 'next/server'
import { z } from 'zod'
import {
  checkEmailRateLimit,
  checkIPRateLimit,
  recordEmailRequest,
  recordIPRequest,
} from '@/lib/console/rate-limiter'
import { createMagicLinkToken } from '@/lib/console/storage/magic-links'
import { sendMagicLinkEmail } from '@/lib/console/email/sendgrid'

// Email validation schema
const requestSchema = z.object({
  email: z.string().email('Invalid email address'),
  captchaToken: z.string().min(1, 'Captcha verification required'),
})

/**
 * Verify Cloudflare Turnstile captcha
 */
async function verifyCaptcha(token: string, ip: string): Promise<boolean> {
  const secretKey = process.env.TURNSTILE_SECRET_KEY

  if (!secretKey) {
    console.error('TURNSTILE_SECRET_KEY is not configured')
    return false
  }

  try {
    const response = await fetch(
      'https://challenges.cloudflare.com/turnstile/v0/siteverify',
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          secret: secretKey,
          response: token,
          remoteip: ip,
        }),
      }
    )

    const data = await response.json()
    return data.success === true
  } catch (error) {
    console.error('Captcha verification error:', error)
    return false
  }
}

/**
 * Get client IP address from request
 */
function getClientIP(request: NextRequest): string {
  const forwardedFor = request.headers.get('x-forwarded-for')
  if (forwardedFor) {
    return forwardedFor.split(',')[0].trim()
  }

  const realIP = request.headers.get('x-real-ip')
  if (realIP) {
    return realIP
  }

  return 'unknown'
}

export async function POST(request: NextRequest) {
  try {
    // Parse and validate request body
    const body = await request.json()
    const validation = requestSchema.safeParse(body)

    if (!validation.success) {
      return NextResponse.json(
        { error: 'Invalid request', details: validation.error.errors },
        { status: 400 }
      )
    }

    const { email, captchaToken } = validation.data
    const ip = getClientIP(request)
    const userAgent = request.headers.get('user-agent') || 'unknown'

    // Verify captcha (skip in development mode)
    const isDevelopment = process.env.NODE_ENV === 'development'
    if (!isDevelopment) {
      const captchaValid = await verifyCaptcha(captchaToken, ip)
      if (!captchaValid) {
        return NextResponse.json(
          { error: 'Captcha verification failed' },
          { status: 400 }
        )
      }
    }

    // Check rate limits
    const emailAllowed = checkEmailRateLimit(email)
    const ipAllowed = checkIPRateLimit(ip)

    if (!emailAllowed || !ipAllowed) {
      // Return success even if rate limited (security: don't reveal rate limiting)
      // But don't actually send the email
      console.warn(`Rate limit exceeded - Email: ${email}, IP: ${ip}`)
      return NextResponse.json({
        success: true,
        message: 'If this email is registered, you will receive a magic link shortly.',
      })
    }

    // Record the request
    recordEmailRequest(email)
    recordIPRequest(ip)

    // Create magic link token
    const token = await createMagicLinkToken(email, ip, userAgent)

    // Generate magic link URL
    const baseUrl = process.env.NEXT_PUBLIC_CONSOLE_URL || request.nextUrl.origin
    const magicLink = `${baseUrl}/api/auth/magic-link/verify?token=${token}`

    // Send email
    try {
      await sendMagicLinkEmail(email, magicLink)
    } catch (error) {
      console.error('Failed to send magic link email:', error)
      // Return success anyway (security: don't reveal if email failed)
      // In production, you might want to queue for retry
    }

    return NextResponse.json({
      success: true,
      message: 'If this email is registered, you will receive a magic link shortly.',
    })
  } catch (error) {
    console.error('Magic link send error:', error)
    return NextResponse.json(
      { error: 'An error occurred. Please try again.' },
      { status: 500 }
    )
  }
}
