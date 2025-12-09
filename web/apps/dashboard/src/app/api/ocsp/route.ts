import { NextRequest, NextResponse } from 'next/server'
import { createOCSPResponder, OCSPResponder } from '@/lib/ocsp'

let ocspResponder: OCSPResponder | null = null

async function getOCSPResponder(): Promise<OCSPResponder> {
  if (ocspResponder) {
    return ocspResponder
  }

  // Get CA cert and key from environment variables
  // These should be mounted from the Kubernetes secret
  const caCertPem = process.env.CA_CERT_PEM
  const caKeyPem = process.env.CA_KEY_PEM

  if (!caCertPem || !caKeyPem) {
    throw new Error('CA_CERT_PEM and CA_KEY_PEM environment variables are required')
  }

  // Decode base64 if needed (for Kubernetes secrets mounted as env vars)
  const decodedCaCert = caCertPem.includes('-----BEGIN')
    ? caCertPem
    : Buffer.from(caCertPem, 'base64').toString('utf-8')

  const decodedCaKey = caKeyPem.includes('-----BEGIN')
    ? caKeyPem
    : Buffer.from(caKeyPem, 'base64').toString('utf-8')

  ocspResponder = await createOCSPResponder(decodedCaCert, decodedCaKey)
  return ocspResponder
}

export async function POST(request: NextRequest) {
  try {
    const responder = await getOCSPResponder()

    // Read the binary OCSP request
    const requestBody = await request.arrayBuffer()

    // Process the OCSP request
    const responseBytes = await responder.handleOCSPRequest(requestBody)

    // Return the binary OCSP response
    return new NextResponse(Buffer.from(responseBytes), {
      status: 200,
      headers: {
        'Content-Type': 'application/ocsp-response',
        'Cache-Control': 'max-age=3600', // Cache for 1 hour
      },
    })
  } catch (error) {
    console.error('OCSP request error:', error)

    // Return an internal error OCSP response
    // Response status 2 = internalError
    const errorResponse = new Uint8Array([0x30, 0x03, 0x0a, 0x01, 0x02])
    return new NextResponse(Buffer.from(errorResponse), {
      status: 200,
      headers: {
        'Content-Type': 'application/ocsp-response',
      },
    })
  }
}

// OCSP also supports GET requests with base64-encoded request in URL path
export async function GET(request: NextRequest) {
  try {
    const responder = await getOCSPResponder()

    // Get the base64-encoded OCSP request from the URL path
    // The request is typically appended to the URL path after /api/ocsp/
    const url = new URL(request.url)
    const pathParts = url.pathname.split('/api/ocsp/')

    if (pathParts.length < 2 || !pathParts[1]) {
      return new NextResponse('OCSP request required', { status: 400 })
    }

    // URL-decode and base64-decode the request
    const encodedRequest = decodeURIComponent(pathParts[1])
    const requestBytes = Buffer.from(encodedRequest, 'base64')

    // Process the OCSP request
    const responseBytes = await responder.handleOCSPRequest(requestBytes.buffer.slice(
      requestBytes.byteOffset,
      requestBytes.byteOffset + requestBytes.byteLength
    ))

    // Return the binary OCSP response
    return new NextResponse(Buffer.from(responseBytes), {
      status: 200,
      headers: {
        'Content-Type': 'application/ocsp-response',
        'Cache-Control': 'max-age=3600', // Cache for 1 hour
      },
    })
  } catch (error) {
    console.error('OCSP GET request error:', error)

    // Return an internal error OCSP response
    const errorResponse = new Uint8Array([0x30, 0x03, 0x0a, 0x01, 0x02])
    return new NextResponse(Buffer.from(errorResponse), {
      status: 200,
      headers: {
        'Content-Type': 'application/ocsp-response',
      },
    })
  }
}
