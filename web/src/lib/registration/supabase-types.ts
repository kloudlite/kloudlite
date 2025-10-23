/**
 * Supabase Database Type Definitions
 */

export type Database = {
  public: {
    Tables: {
      user_registrations: {
        Row: {
          email: string
          user_id: string
          name: string
          providers: ('github' | 'google' | 'azure-ad')[]
          registered_at: string
          installation_key: string
          secret_key: string | null
          has_completed_installation: boolean
          subdomain: string | null
          reserved_at: string | null
          deployment_ready: boolean | null
          last_health_check: string | null
          created_at: string
          updated_at: string
        }
        Insert: {
          email: string
          user_id: string
          name: string
          providers?: ('github' | 'google' | 'azure-ad')[]
          registered_at?: string
          installation_key: string
          secret_key?: string | null
          has_completed_installation?: boolean
          subdomain?: string | null
          reserved_at?: string | null
          deployment_ready?: boolean | null
          last_health_check?: string | null
        }
        Update: {
          email?: string
          user_id?: string
          name?: string
          providers?: ('github' | 'google' | 'azure-ad')[]
          registered_at?: string
          installation_key?: string
          secret_key?: string | null
          has_completed_installation?: boolean
          subdomain?: string | null
          reserved_at?: string | null
          deployment_ready?: boolean | null
          last_health_check?: string | null
        }
      }
      ip_records: {
        Row: {
          id: number
          user_email: string
          type: 'installation' | 'workmachine'
          ip: string
          work_machine_name: string | null
          configured_at: string
          dns_record_ids: string[]
          created_at: string
          updated_at: string
        }
        Insert: {
          user_email: string
          type: 'installation' | 'workmachine'
          ip: string
          work_machine_name?: string | null
          configured_at?: string
          dns_record_ids?: string[]
        }
        Update: {
          user_email?: string
          type?: 'installation' | 'workmachine'
          ip?: string
          work_machine_name?: string | null
          configured_at?: string
          dns_record_ids?: string[]
        }
      }
      domain_reservations: {
        Row: {
          subdomain: string
          user_id: string
          user_email: string
          user_name: string
          reserved_at: string
          status: 'reserved' | 'active' | 'cancelled'
          created_at: string
          updated_at: string
        }
        Insert: {
          subdomain: string
          user_id: string
          user_email: string
          user_name: string
          reserved_at?: string
          status?: 'reserved' | 'active' | 'cancelled'
        }
        Update: {
          subdomain?: string
          user_id?: string
          user_email?: string
          user_name?: string
          reserved_at?: string
          status?: 'reserved' | 'active' | 'cancelled'
        }
      }
      tls_certificates: {
        Row: {
          id: number
          user_email: string
          cloudflare_cert_id: string | null
          certificate: string
          private_key: string
          hostnames: string[]
          scope: 'installation' | 'workmachine' | 'workspace'
          scope_identifier: string | null
          parent_scope_identifier: string | null
          valid_from: string
          valid_until: string
          generated_at: string
          created_at: string
          updated_at: string
        }
        Insert: {
          user_email: string
          cloudflare_cert_id?: string | null
          certificate: string
          private_key: string
          hostnames: string[]
          scope: 'installation' | 'workmachine' | 'workspace'
          scope_identifier?: string | null
          parent_scope_identifier?: string | null
          valid_from: string
          valid_until: string
          generated_at?: string
        }
        Update: {
          user_email?: string
          cloudflare_cert_id?: string | null
          certificate?: string
          private_key?: string
          hostnames?: string[]
          scope?: 'installation' | 'workmachine' | 'workspace'
          scope_identifier?: string | null
          parent_scope_identifier?: string | null
          valid_from?: string
          valid_until?: string
          generated_at?: string
        }
      }
    }
  }
}
