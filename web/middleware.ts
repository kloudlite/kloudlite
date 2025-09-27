import { withAuth } from "next-auth/middleware"
import { NextResponse } from "next/server"

export default withAuth(
  function middleware(req) {
    const token = req.nextauth.token
    const path = req.nextUrl.pathname

    // Public routes that don't require authentication
    const publicRoutes = [
      "/",
      "/auth/login",
      "/auth/signup", 
      "/auth/forgot-password",
      "/auth/reset-password",
      "/auth/verify-email",
      "/auth/email-verification-required",
      "/device",
    ]

    // Routes that require authentication but not email verification
    const authOnlyRoutes = [
      "/auth/email-verification-required",
    ]

    // Check if the current path is public
    if (publicRoutes.includes(path)) {
      return NextResponse.next()
    }

    // If user is not authenticated, redirect to login
    if (!token) {
      const loginUrl = new URL("/auth/login", req.url)
      loginUrl.searchParams.set("callbackUrl", path)
      return NextResponse.redirect(loginUrl)
    }

    // Check if email is verified for protected routes
    if (!authOnlyRoutes.includes(path) && !token.emailVerified) {
      // Redirect unverified users to email verification required page
      return NextResponse.redirect(new URL("/auth/email-verification-required", req.url))
    }

    // User is authenticated and verified (or on auth-only route)
    return NextResponse.next()
  },
  {
    callbacks: {
      authorized: ({ token }) => {
        // This callback is called before the middleware function above
        // We allow all requests to pass through here and handle the logic above
        return true
      },
    },
    pages: {
      signIn: "/auth/login",
      error: "/auth/error",
    },
  }
)

// Configure which routes to protect
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api/auth (auth API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public (public files)
     */
    "/((?!api/auth|_next/static|_next/image|favicon.ico|public).*)",
  ],
}