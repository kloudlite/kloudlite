import { NextRequest, NextResponse } from 'next/server'
import { cookies } from 'next/headers'

export async function POST(request: NextRequest) {
  const cookieStore = await cookies()

  // Clear the registration session cookie
  cookieStore.delete('registration_session')

  return NextResponse.json({ success: true })
}
