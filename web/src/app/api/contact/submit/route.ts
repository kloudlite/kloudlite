import { NextRequest, NextResponse } from 'next/server'
import { saveContactSubmission } from '@/lib/console/supabase-storage-service'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'

/**
 * Submit contact form
 */
export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { name, email, subject, message } = body

    // Validate required fields
    if (!name || !email || !subject || !message) {
      return NextResponse.json(
        { error: 'All fields are required' },
        { status: 400 }
      )
    }

    // Basic email validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(email)) {
      return NextResponse.json(
        { error: 'Invalid email address' },
        { status: 400 }
      )
    }

    // Validate field lengths
    if (name.length < 2 || name.length > 100) {
      return NextResponse.json(
        { error: 'Name must be between 2 and 100 characters' },
        { status: 400 }
      )
    }

    if (subject.length < 3 || subject.length > 200) {
      return NextResponse.json(
        { error: 'Subject must be between 3 and 200 characters' },
        { status: 400 }
      )
    }

    if (message.length < 10 || message.length > 5000) {
      return NextResponse.json(
        { error: 'Message must be between 10 and 5000 characters' },
        { status: 400 }
      )
    }

    // Save to Supabase
    const submission = await saveContactSubmission({
      name: name.trim(),
      email: email.trim().toLowerCase(),
      subject: subject.trim(),
      message: message.trim(),
    })

    console.log('Contact form submission saved:', submission.id)

    return NextResponse.json({
      success: true,
      message: 'Thank you for contacting us! We will get back to you soon.',
      id: submission.id,
    })
  } catch (error) {
    console.error('Contact form submission error:', error)
    return NextResponse.json(
      { error: 'Failed to submit contact form. Please try again later.' },
      { status: 500 }
    )
  }
}
