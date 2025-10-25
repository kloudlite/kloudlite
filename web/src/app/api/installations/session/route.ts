import { NextResponse } from 'next/server'
import { cookies } from 'next/headers'
import { jwtVerify } from 'jose'

export async function GET() {
  const cookieStore = await cookies()
  const token = cookieStore.get('registration_session')?.value

  if (!token) {
    return NextResponse.json({ error: 'Not authenticated' }, { status: 401 })
  }

  try {
    const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
    const { payload } = await jwtVerify(token, secret)

    const response = NextResponse.json({
      user: {
        email: payload.email as string,
        name: payload.name as string,
        image: payload.image as string,
      },
      provider: payload.provider,
      installationKey: payload.installationKey as string,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch {
    return NextResponse.json({ error: 'Invalid session' }, { status: 401 })
  }
}
