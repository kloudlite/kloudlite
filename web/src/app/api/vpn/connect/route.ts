import { NextRequest, NextResponse } from 'next/server'

/**
 * VPN Connect API Route
 * Proxies VPN connection requests to the backend API server
 * Used by kltun CLI for establishing VPN connections
 */
export async function GET(request: NextRequest) {
  try {
    // Get the authorization header from the incoming request
    const authHeader = request.headers.get('authorization')

    if (!authHeader) {
      return NextResponse.json(
        { error: 'Authorization header required' },
        { status: 401 }
      )
    }

    // Get the backend API URL from environment
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

    // Forward the request to the backend Go API
    const backendResponse = await fetch(`${backendUrl}/api/v1/vpn/connect`, {
      method: 'GET',
      headers: {
        'Authorization': authHeader,
        'Content-Type': 'application/json',
      },
    })

    // Get the response data
    const data = await backendResponse.json()

    // Return the backend response with the same status code
    return NextResponse.json(data, { status: backendResponse.status })
  } catch (error) {
    console.error('VPN connect proxy error:', error)
    return NextResponse.json(
      { error: 'Failed to connect to VPN service' },
      { status: 500 }
    )
  }
}
