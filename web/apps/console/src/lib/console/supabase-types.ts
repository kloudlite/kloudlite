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
          providers: ('github' | 'google' | 'azure-ad' | 'email')[]
          registered_at: string
          created_at: string
          updated_at: string
        }
        Insert: {
          user_id: string
          email: string
          name: string
          providers?: ('github' | 'google' | 'azure-ad' | 'email')[]
          registered_at?: string
        }
        Update: {
          email?: string
          name?: string
          providers?: ('github' | 'google' | 'azure-ad' | 'email')[]
          registered_at?: string
        }
      }
      magic_link_tokens: {
        Row: {
          id: string
          email: string
          token: string
          expires_at: string
          used_at: string | null
          created_at: string
          ip_address: string | null
          user_agent: string | null
        }
        Insert: {
          id?: string
          email: string
          token: string
          expires_at: string
          used_at?: string | null
          created_at?: string
          ip_address?: string | null
          user_agent?: string | null
        }
        Update: {
          used_at?: string | null
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
          cloud_provider: 'aws' | 'gcp' | 'azure' | 'oci' | null
          cloud_location: string | null
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
          cloud_provider?: 'aws' | 'gcp' | 'azure' | 'oci' | null
          cloud_location?: string | null
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
          cloud_provider?: 'aws' | 'gcp' | 'azure' | 'oci' | null
          cloud_location?: string | null
        }
      }
      ip_records: {
        Row: {
          id: number
          installation_id: string
          domain_request_name: string
          ip: string
          configured_at: string
          ssh_record_id: string | null
          route_record_ids: string[]
          route_record_map: Record<string, string> | null
          domain_routes: Array<{ domain: string }>
          created_at: string
          updated_at: string
        }
        Insert: {
          installation_id: string
          domain_request_name: string
          ip: string
          configured_at?: string
          ssh_record_id?: string | null
          route_record_ids?: string[]
          route_record_map?: Record<string, string> | null
          domain_routes?: Array<{ domain: string }>
        }
        Update: {
          installation_id?: string
          domain_request_name?: string
          ip?: string
          configured_at?: string
          ssh_record_id?: string | null
          route_record_ids?: string[]
          route_record_map?: Record<string, string> | null
          domain_routes?: Array<{ domain: string }>
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
      installation_members: {
        Row: {
          id: string
          installation_id: string
          user_id: string
          role: 'owner' | 'admin' | 'member' | 'viewer'
          added_by: string | null
          added_at: string
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          installation_id: string
          user_id: string
          role: 'owner' | 'admin' | 'member' | 'viewer'
          added_by?: string | null
          added_at?: string
        }
        Update: {
          role?: 'owner' | 'admin' | 'member' | 'viewer'
        }
      }
      installation_invitations: {
        Row: {
          id: string
          installation_id: string
          email: string
          role: 'admin' | 'member' | 'viewer'
          invited_by: string
          status: 'pending' | 'accepted' | 'rejected' | 'expired'
          expires_at: string
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          installation_id: string
          email: string
          role: 'admin' | 'member' | 'viewer'
          invited_by: string
          status?: 'pending' | 'accepted' | 'rejected' | 'expired'
          expires_at?: string
        }
        Update: {
          status?: 'pending' | 'accepted' | 'rejected' | 'expired'
          role?: 'admin' | 'member' | 'viewer'
        }
      }
    }
  }
}
