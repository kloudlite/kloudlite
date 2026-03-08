/**
 * DNS Configuration Management
 */

import { supabase } from '../supabase'
import type { DnsConfiguration, DnsConfigurationRow } from './types'

/**
 * Atomically add or update DNS configuration
 */
export async function upsertDnsConfiguration(
  installationId: string,
  dnsConfig: DnsConfiguration,
): Promise<number> {
  const { error } = await supabase
    .from('dns_configurations')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .upsert(
      {
        installation_id: installationId,
        service_name: dnsConfig.serviceName,
        ip: dnsConfig.ip,
        ssh_record_id: dnsConfig.sshRecordId || null,
        route_record_ids: dnsConfig.routeRecordIds || [],
        route_record_map: dnsConfig.routeRecordMap || {},
        domain_routes: dnsConfig.domainRoutes || [],
      },
      {
        onConflict: 'installation_id,service_name',
      },
    )

  if (error) {
    throw new Error(`Failed to add DNS configuration: ${error.message}`)
  }

  // Get total count
  const { count } = await supabase
    .from('dns_configurations')
    .select('*', { count: 'exact', head: true })
    .eq('installation_id', installationId)

  return count || 0
}

/**
 * Remove a specific DNS configuration for a service
 */
export async function removeDnsConfiguration(
  installationId: string,
  serviceName: string,
): Promise<void> {
  const { error } = await supabase
    .from('dns_configurations')
    .delete()
    .eq('installation_id', installationId)
    .eq('service_name', serviceName)

  if (error) {
    console.error('Error removing DNS configuration:', error)
    throw new Error(`Failed to remove DNS configuration: ${error.message}`)
  }
}

/**
 * Delete all DNS configurations for an installation and return their DNS record IDs
 */
export async function deleteDnsConfigurations(installationId: string): Promise<string[]> {
  // Get all DNS configurations to extract DNS record IDs
  const { data: configData } = await supabase
    .from('dns_configurations')
    .select('*')
    .eq('installation_id', installationId)

  const dnsRecordIds: string[] = []

  if (configData) {
    for (const record of configData as DnsConfigurationRow[]) {
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

  // Delete all DNS configurations
  const { error } = await supabase
    .from('dns_configurations')
    .delete()
    .eq('installation_id', installationId)

  if (error) {
    console.error('Error deleting DNS configurations:', error)
  }

  return dnsRecordIds
}
