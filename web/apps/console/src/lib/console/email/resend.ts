/**
 * Resend email service for magic link delivery
 * Uses Resend API (https://resend.com)
 */

import { generateMagicLinkEmail } from './templates/magic-link'

const RESEND_API_KEY = process.env.RESEND_API_KEY
const RESEND_FROM_EMAIL = process.env.RESEND_FROM_EMAIL || 'noreply@kloudlite.io'
const RESEND_FROM_NAME = process.env.RESEND_FROM_NAME || 'Kloudlite'

/**
 * Send a magic link email via Resend
 * @param email - Recipient email address
 * @param magicLink - Full magic link URL
 * @throws Error if Resend API request fails
 */
export async function sendMagicLinkEmail(
  email: string,
  magicLink: string
): Promise<void> {
  console.log('Resend API Key exists:', !!RESEND_API_KEY)
  console.log('From email:', RESEND_FROM_EMAIL)

  if (!RESEND_API_KEY) {
    throw new Error('RESEND_API_KEY is not configured')
  }

  const { html, text } = generateMagicLinkEmail(magicLink)

  const response = await fetch('https://api.resend.com/emails', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${RESEND_API_KEY}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      from: `${RESEND_FROM_NAME} <${RESEND_FROM_EMAIL}>`,
      to: [email],
      subject: 'Sign in to Kloudlite Console',
      html,
      text,
    }),
  })

  if (!response.ok) {
    const errorText = await response.text()
    console.error('Resend API error:', response.status, errorText)
    throw new Error(`Failed to send email: ${response.status}`)
  }
}
