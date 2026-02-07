import { NextRequest, NextResponse } from 'next/server'
import { signIn } from '@/lib/auth'

export async function GET(request: NextRequest) {
  const token = request.nextUrl.searchParams.get('token')

  if (!token) {
    return NextResponse.redirect(new URL('/superadmin-login?error=missing', request.url))
  }

  try {
    await signIn('credentials', {
      superadminToken: token,
      redirectTo: '/admin',
    })
  } catch (error: any) {
    // NextAuth v5 signIn throws a NEXT_REDIRECT on success — re-throw it
    if (error?.digest?.startsWith('NEXT_REDIRECT')) {
      throw error
    }
    // Auth failed — redirect back to login page with error
    return NextResponse.redirect(
      new URL(`/superadmin-login?token=${encodeURIComponent(token)}&error=1`, request.url)
    )
  }

  // Fallback (shouldn't reach here)
  return NextResponse.redirect(new URL('/admin', request.url))
}
