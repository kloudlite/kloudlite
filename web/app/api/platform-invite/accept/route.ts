import { Metadata } from '@grpc/grpc-js'
import { type NextRequest, NextResponse } from 'next/server'

import { getAccountsClient } from '@/lib/grpc/accounts-client'

export async function POST(request: NextRequest) {
  try {
    const { token } = await request.json()

    if (!token) {
      return NextResponse.json(
        { success: false, error: 'Token is required' },
        { status: 400 }
      )
    }

    const client = getAccountsClient()
    const metadata = new Metadata()

    return new Promise((resolve) => {
      client.acceptPlatformInvitation(
        { token },
        metadata,
        (error, response) => {
          if (error) {
            resolve(
              NextResponse.json(
                { success: false, error: error.message },
                { status: 400 }
              )
            )
          } else if (response?.success) {
            resolve(
              NextResponse.json({ success: true })
            )
          } else {
            resolve(
              NextResponse.json(
                { success: false, error: response?.error || 'Failed to accept invitation' },
                { status: 400 }
              )
            )
          }
        }
      )
    })
  } catch (_error) {
    // Error accepting platform invitation
    return NextResponse.json(
      { success: false, error: 'Internal server error' },
      { status: 500 }
    )
  }
}