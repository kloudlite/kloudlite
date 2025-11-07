import { NextRequest, NextResponse } from 'next/server'
import {
  getInstallationByKey,
  deleteEdgeCertificatesForDomainRequest,
} from '@/lib/console/supabase-storage-service'
import { supabase } from '@/lib/console/supabase'
import { revokeCertificate } from '@/lib/console/cloudflare-certificates'

// Use Node.js runtime for Supabase
export const runtime = 'nodejs'

/**
 * Delete all certificates for a DomainRequest
 * Called by the DomainRequest controller finalizer when DomainRequest is deleted
 *
 * Deletes from both Supabase and Cloudflare:
 * 1. Origin certificates (Cloudflare Origin CA) - stored with key (installationId, scope='workmachine', scopeIdentifier=domainRequestName)
 * 2. Edge certificates (Cloudflare Edge TLS) - stored with key (installationId, domainRequestName)
 *
 * Note: DNS records are deleted separately via the configure-ips endpoint
 *
 * Request format (JSON body):
 * {
 *   "installationKey": "abc-123",
 *   "domainRequestName": "wm-karthik"
 * }
 *
 * Requires Authorization: Bearer <secret-key>
 */
export async function POST(request: NextRequest) {
  try {
    // Extract and validate bearer token
    const authHeader = request.headers.get('authorization')
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return NextResponse.json(
        { error: 'Missing or invalid authorization header' },
        { status: 401 },
      )
    }

    const secretKey = authHeader.substring(7) // Remove "Bearer " prefix

    // Parse request body
    const body = await request.json()
    const installationKey = body.installationKey
    const domainRequestName = body.domainRequestName as string | undefined

    if (!installationKey) {
      return NextResponse.json({ error: 'Installation key is required' }, { status: 400 })
    }

    if (!domainRequestName) {
      return NextResponse.json({ error: 'domainRequestName is required' }, { status: 400 })
    }

    // Look up installation by installation key
    const installation = await getInstallationByKey(installationKey)

    if (!installation) {
      return NextResponse.json({ error: 'Invalid installation key' }, { status: 404 })
    }

    // Verify secret key matches
    if (installation.secretKey !== secretKey) {
      return NextResponse.json({ error: 'Invalid secret key' }, { status: 403 })
    }

    console.log(
      `Deleting all resources for domainRequest: ${domainRequestName}, installation: ${installation.id}`,
    )

    // 1. Delete Origin Certificates from Cloudflare and Supabase
    const { data: originCerts } = await supabase
      .from('tls_certificates')
      .select('cloudflare_cert_id')
      .eq('installation_id', installation.id)
      .eq('scope', 'workmachine')
      .eq('scope_identifier', domainRequestName)

    if (originCerts && originCerts.length > 0) {
      for (const cert of originCerts) {
        if (cert.cloudflare_cert_id) {
          try {
            await revokeCertificate(cert.cloudflare_cert_id)
            console.log(`Revoked origin certificate from Cloudflare: ${cert.cloudflare_cert_id}`)
          } catch (error) {
            console.error(`Failed to revoke origin certificate: ${cert.cloudflare_cert_id}`, error)
            // Continue with deletion even if revocation fails
          }
        }
      }

      // Delete from Supabase
      const { error } = await supabase
        .from('tls_certificates')
        .delete()
        .eq('installation_id', installation.id)
        .eq('scope', 'workmachine')
        .eq('scope_identifier', domainRequestName)

      if (error) {
        console.error('Error deleting origin certificates from Supabase:', error)
      } else {
        console.log(`Deleted ${originCerts.length} origin certificate(s) from Supabase`)
      }
    }

    // 2. Delete Edge Certificates from Cloudflare and Supabase
    const deletedEdgeCertIds = await deleteEdgeCertificatesForDomainRequest(
      installation.id,
      domainRequestName,
    )
    console.log(`Deleted ${deletedEdgeCertIds.length} edge certificate(s)`)

    // Note: DNS records are deleted by the configure-ips endpoint (called separately by controller)

    return NextResponse.json({
      success: true,
      message: 'All DomainRequest certificates deleted successfully',
      deletedResources: {
        originCertificates: originCerts?.length || 0,
        edgeCertificates: deletedEdgeCertIds.length,
      },
    })
  } catch (error) {
    console.error('Delete DomainRequest resources error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
