/**
 * IP Record Management
 */

import { supabase } from '../supabase'
import type { IPRecord, IPRecordRow } from './types'

/**
 * Atomically add or update IP record
 */
export async function addOrUpdateIpRecord(
  installationId: string,
  ipRecord: IPRecord,
): Promise<number> {
  const { error } = await supabase
    .from('ip_records')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .upsert(
      {
        installation_id: installationId,
        domain_request_name: ipRecord.domainRequestName,
        ip: ipRecord.ip,
        configured_at: ipRecord.configuredAt,
        ssh_record_id: ipRecord.sshRecordId || null,
        route_record_ids: ipRecord.routeRecordIds || [],
        route_record_map: ipRecord.routeRecordMap || {},
        domain_routes: ipRecord.domainRoutes || [],
      },
      {
        onConflict: 'installation_id,domain_request_name',
      },
    )

  if (error) {
    throw new Error(`Failed to add IP record: ${error.message}`)
  }

  // Get total count
  const { count } = await supabase
    .from('ip_records')
    .select('*', { count: 'exact', head: true })
    .eq('installation_id', installationId)

  return count || 0
}

/**
 * Remove a specific IP record for a domain request
 */
export async function removeIpRecord(
  installationId: string,
  domainRequestName: string,
): Promise<void> {
  const { error } = await supabase
    .from('ip_records')
    .delete()
    .eq('installation_id', installationId)
    .eq('domain_request_name', domainRequestName)

  if (error) {
    console.error('Error removing IP record:', error)
    throw new Error(`Failed to remove IP record: ${error.message}`)
  }
}

/**
 * Delete all IP records for an installation and return their DNS record IDs
 */
export async function deleteIpRecords(installationId: string): Promise<string[]> {
  // Get all IP records to extract DNS record IDs
  const { data: ipData } = await supabase
    .from('ip_records')
    .select('*')
    .eq('installation_id', installationId)

  const dnsRecordIds: string[] = []

  if (ipData) {
    for (const record of ipData as IPRecordRow[]) {
      // Collect SSH record ID
      if (record.ssh_record_id) {
        dnsRecordIds.push(record.ssh_record_id)
      }
      // Collect route record IDs
      if (record.route_record_ids && Array.isArray(record.route_record_ids)) {
        dnsRecordIds.push(...record.route_record_ids)
      }
    }
  }

  // Delete all IP records
  const { error } = await supabase.from('ip_records').delete().eq('installation_id', installationId)

  if (error) {
    console.error('Error deleting IP records:', error)
  }

  return dnsRecordIds
}
