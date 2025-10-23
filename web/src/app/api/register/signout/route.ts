import { NextResponse } from 'next/server'
import { cookies } from 'next/headers'

export async function POST() {
  const cookieStore = await cookies()

  // Clear the registration session cookie
  cookieStore.delete('registration_session')

  return NextResponse.json({ success: true })
}
