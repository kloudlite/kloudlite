import { NextRequest, NextResponse } from 'next/server'
import { env } from '@/lib/env'

export async function GET(request: NextRequest) {
  try {
    // Get the authorization header from the request
    const authHeader = request.headers.get('authorization')

    if (!authHeader) {
      return NextResponse.json(
        { error: 'Authorization header required' },
        { status: 401 }
      )
    }

    // Extract the connection token from Bearer header
    const token = authHeader.replace('Bearer ', '')

    // Decode the JWT to get user information (we don't need to verify since backend will verify)
    // For now, we'll just forward the request to the backend

    // Proxy the request to the backend API
    // Since connection tokens can access workspaces across namespaces,
    // we need to list all workspaces the user has access to
    const backendUrl = `${env.apiUrl}/api/v1/namespaces/wm-user/workspaces`

    const response = await fetch(backendUrl, {
      headers: {
        'Authorization': authHeader,
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      const errorText = await response.text()
      return NextResponse.json(
        { error: errorText || 'Failed to fetch workspaces' },
        { status: response.status }
      )
    }

    const data = await response.json()

    // Return the workspaces in the format expected by the VS Code extension
    return NextResponse.json({
      workspaces: data.items || []
    })

  } catch (error: any) {
    console.error('VS Code API error:', error)
    return NextResponse.json(
      { error: error.message || 'Internal server error' },
      { status: 500 }
    )
  }
}
