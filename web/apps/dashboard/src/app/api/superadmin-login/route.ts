import { NextRequest, NextResponse } from 'next/server'
import { signIn } from '@/lib/auth'

function redirectTo(path: string) {
  return new NextResponse(null, {
    status: 303,
    headers: {
      Location: path,
    },
  })
}

export async function GET(request: NextRequest) {
  const token = request.nextUrl.searchParams.get('token')

  if (!token) {
    return redirectTo('/superadmin-login?error=missing')
  }

  try {
    await signIn('credentials', {
      superadminToken: token,
      redirectTo: '/admin',
    })
  } catch (error: unknown) {
    // NextAuth v5 signIn throws a NEXT_REDIRECT on success — re-throw it
    if (
      typeof error === 'object' &&
      error !== null &&
      'digest' in error &&
      typeof (error as { digest?: string }).digest === 'string' &&
      (error as { digest: string }).digest.startsWith('NEXT_REDIRECT')
    ) {
      throw error
    }
    // Auth failed — redirect back to login page with error
    return redirectTo(`/superadmin-login?token=${encodeURIComponent(token)}&error=1`)
  }

  // Fallback (shouldn't reach here)
  return redirectTo('/admin')
}
