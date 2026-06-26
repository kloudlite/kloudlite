import { NextResponse } from 'next/server'

export function apiError(message: string, status: number) {
  return NextResponse.json({ error: message }, { status })
}

export function apiCatchError(error: unknown, fallback: string, defaultStatus = 500) {
  const message = error instanceof Error ? error.message : fallback
  const status = message.includes('Unauthorized') ? 401 : message.includes('Forbidden') ? 403 : defaultStatus
  return NextResponse.json({ error: message }, { status })
}
