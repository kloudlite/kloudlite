import { NextResponse } from 'next/server'
import { getAuthToken } from '@/lib/get-session'

export async function GET() {
  const token = await getAuthToken()

  if (!token) {
    return NextResponse.json({ error: 'Not authenticated' }, { status: 401 })
  }

  return NextResponse.json({ token })
}
