import { NextRequest, NextResponse } from 'next/server'
import { createRequire } from 'module'

// Use Node.js runtime for Supabase
export const runtime = 'nodejs'

const require = createRequire(import.meta.url)

/**
 * Get PII Supabase client for contact_messages table.
 * Contact messages contain PII and live in the dedicated PII database.
 */
function getPiiSupabaseClient() {
  const supabaseUrl = process.env.PII_SUPABASE_URL
  const supabaseKey = process.env.PII_SUPABASE_KEY
  if (!supabaseUrl || !supabaseKey) {
    return null
  }

  const { createClient } = require('@supabase/supabase-js') as typeof import('@supabase/supabase-js')
  return createClient(supabaseUrl, supabaseKey)
}

interface ContactSubmission {
  name: string
  email: string
  subject: string
  message: string
}

export async function POST(request: NextRequest) {
  const supabase = getPiiSupabaseClient()
  if (!supabase) {
    return NextResponse.json(
      { error: 'Contact form is not configured' },
      { status: 500 }
    )
  }

  try {
    const body = await request.json() as ContactSubmission

    // Validate required fields
    if (!body.name || !body.email || !body.subject || !body.message) {
      return NextResponse.json(
        { error: 'All fields are required' },
        { status: 400 }
      )
    }

    // Validate email format
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(body.email)) {
      return NextResponse.json(
        { error: 'Invalid email address' },
        { status: 400 }
      )
    }

    // Get client IP and user agent for audit trail
    const ip = request.headers.get('x-forwarded-for') ||
               request.headers.get('x-real-ip') ||
               'unknown'
    const userAgent = request.headers.get('user-agent') || 'unknown'

    // Insert into PII database
    const { data, error } = await supabase
      .from('contact_messages')
      .insert({
        name: body.name.trim(),
        email: body.email.trim().toLowerCase(),
        subject: body.subject.trim(),
        message: body.message.trim(),
        ip_address: ip,
        user_agent: userAgent,
      })
      .select()
      .single()

    if (error) {
      console.error('Supabase error:', error)
      return NextResponse.json(
        { error: 'Failed to submit contact form' },
        { status: 500 }
      )
    }

    return NextResponse.json({
      message: "Thank you for your message! We'll get back to you soon.",
      id: data.id,
    })
  } catch (error) {
    console.error('Contact form error:', error)
    return NextResponse.json(
      { error: 'An unexpected error occurred' },
      { status: 500 }
    )
  }
}
