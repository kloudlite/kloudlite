import { NextResponse } from 'next/server'
import { getServerSession } from 'next-auth'

import { getAuthOptions } from '@/lib/auth/get-auth-options'
import { getAccountsClient, getAuthMetadata } from '@/lib/grpc/accounts-client'

export async function GET() {
  try {
    const authOpts = await getAuthOptions()
    const session = await getServerSession(authOpts)
    
    if (!session?.user) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const client = getAccountsClient()
    const metadata = await getAuthMetadata()

    const teams = await new Promise((resolve, reject) => {
      client.listTeams(
        { userId: session.user.id },
        metadata,
        (error, response) => {
          if (error) {
            reject(error)
          } else {
            // Ensure each team has required fields including slug
            const teamsWithSlug = (response?.teams || []).map((team: any) => ({
              ...team,
              accountid: team.teamId,
              slug: team.slug || team.displayName?.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '')
            }))
            resolve(teamsWithSlug)
          }
        }
      )
    })

    return NextResponse.json({ teams })
  } catch (error) {
    console.error('Error fetching teams:', error)
    return NextResponse.json({ teams: [] })
  }
}