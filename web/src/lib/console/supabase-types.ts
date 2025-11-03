/**
 * Supabase Database Type Definitions
 */

export type Database = {
  public: {
    Tables: {
      user_registrations: {
        Row: {
          user_id: string
          email: string
          name: string
          providers: ('github' | 'google' | 'azure-ad')[]
          registered_at: string
          created_at: string
          updated_at: string
        }
        Insert: {
          user_id: string
          email: string
          name: string
          providers?: ('github' | 'google' | 'azure-ad')[]
          registered_at?: string
        }
        Update: {
          email?: string
          name?: string
          providers?: ('github' | 'google' | 'azure-ad')[]
          registered_at?: string
        }
      }
      installations: {
        Row: {
          id: string
          user_id: string
          name: string | null
          description: string | null
          installation_key: string
          secret_key: string | null
          has_completed_installation: boolean
          subdomain: string | null
          reserved_at: string | null
          deployment_ready: boolean | null
          last_health_check: string | null
          edge_certificate_pack_id: string | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          user_id: string
          name?: string | null
          description?: string | null
          installation_key: string
          secret_key?: string | null
          has_completed_installation?: boolean
          subdomain?: string | null
          reserved_at?: string | null
          deployment_ready?: boolean | null
          last_health_check?: string | null
          edge_certificate_pack_id?: string | null
        }
        Update: {
          user_id?: string
          name?: string | null
          description?: string | null
          installation_key?: string
          secret_key?: string | null
          has_completed_installation?: boolean
          subdomain?: string | null
          reserved_at?: string | null
          deployment_ready?: boolean | null
          last_health_check?: string | null
          edge_certificate_pack_id?: string | null
        }
      }
      ip_records: {
        Row: {
          id: number
          installation_id: string
          type: 'installation' | 'workmachine'
          ip: string
          work_machine_name: string | null
          configured_at: string
          dns_record_ids: string[]
          created_at: string
          updated_at: string
        }
        Insert: {
          installation_id: string
          type: 'installation' | 'workmachine'
          ip: string
          work_machine_name?: string | null
          configured_at?: string
          dns_record_ids?: string[]
        }
        Update: {
          installation_id?: string
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
          installation_id: string
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
          installation_id: string
          user_id: string
          user_email: string
          user_name: string
          reserved_at?: string
          status?: 'reserved' | 'active' | 'cancelled'
        }
        Update: {
          subdomain?: string
          installation_id?: string
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
          installation_id: string
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
          installation_id: string
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
          installation_id?: string
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
