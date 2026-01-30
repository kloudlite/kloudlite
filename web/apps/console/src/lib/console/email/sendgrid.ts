/**
 * SendGrid email service for magic link delivery
 * Uses direct REST API integration
 */

import { generateMagicLinkEmail } from './templates/magic-link'

const SENDGRID_API_KEY = process.env.SENDGRID_API_KEY
const SENDGRID_FROM_EMAIL = process.env.SENDGRID_FROM_EMAIL || 'noreply@kloudlite.io'
const SENDGRID_FROM_NAME = process.env.SENDGRID_FROM_NAME || 'Kloudlite'

/**
 * Send a magic link email via SendGrid
 * @param email - Recipient email address
 * @param magicLink - Full magic link URL
 * @throws Error if SendGrid API request fails
 */
export async function sendMagicLinkEmail(
  email: string,
  magicLink: string
): Promise<void> {
  console.log('SendGrid API Key exists:', !!SENDGRID_API_KEY)
  console.log('SendGrid API Key prefix:', SENDGRID_API_KEY?.substring(0, 10))
  console.log('From email:', SENDGRID_FROM_EMAIL)

  if (!SENDGRID_API_KEY) {
    throw new Error('SENDGRID_API_KEY is not configured')
  }

  const { html, text } = generateMagicLinkEmail(magicLink)

  const response = await fetch('https://api.sendgrid.com/v3/mail/send', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${SENDGRID_API_KEY}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      personalizations: [
        {
          to: [{ email }],
        },
      ],
      from: {
        email: SENDGRID_FROM_EMAIL,
        name: SENDGRID_FROM_NAME,
      },
      subject: 'Sign in to Kloudlite Console',
      content: [
        {
          type: 'text/plain',
          value: text,
        },
        {
          type: 'text/html',
          value: html,
        },
      ],
    }),
  })

  if (!response.ok) {
    const errorText = await response.text()
    console.error('SendGrid API error:', response.status, errorText)
    throw new Error(`Failed to send email: ${response.status}`)
  }
}
