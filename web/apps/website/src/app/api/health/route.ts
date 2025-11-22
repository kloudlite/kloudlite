import { NextResponse } from 'next/server'

/**
 * Health check endpoint for Kubernetes readiness/liveness probes
 * Returns 200 OK with minimal processing
 */
export async function GET() {
  return NextResponse.json({ status: 'ok' }, { status: 200 })
}
