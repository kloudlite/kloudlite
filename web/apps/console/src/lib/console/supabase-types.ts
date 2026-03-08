/**
 * Supabase Database Type Definitions (Main DB — operational data)
 * PII tables (users, magic_link_tokens, contact_messages) are in supabase-pii-types.ts
 */

export type Database = {
  public: {
    Tables: {
      organizations: {
        Row: {
          id: string
          name: string
          slug: string
          created_by: string
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          name: string
          slug: string
          created_by: string
        }
        Update: {
          name?: string
          slug?: string
        }
      }
      organization_members: {
        Row: {
          id: string
          org_id: string
          user_id: string
          role: 'owner' | 'admin'
          added_by: string | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          org_id: string
          user_id: string
          role: 'owner' | 'admin'
          added_by?: string | null
        }
        Update: {
          role?: 'owner' | 'admin'
        }
      }
      organization_invitations: {
        Row: {
          id: string
          org_id: string
          email: string
          role: 'admin'
          invited_by: string
          status: 'pending' | 'accepted' | 'rejected' | 'expired'
          expires_at: string
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          org_id: string
          email: string
          role: 'admin'
          invited_by: string
          status?: 'pending' | 'accepted' | 'rejected' | 'expired'
          expires_at?: string
        }
        Update: {
          status?: 'pending' | 'accepted' | 'rejected' | 'expired'
          role?: 'admin'
        }
      }
      installations: {
        Row: {
          id: string
          org_id: string
          name: string | null
          description: string | null
          installation_key: string
          secret_key: string | null
          setup_completed: boolean
          subdomain: string | null
          reserved_at: string | null
          deployment_ready: boolean | null
          last_health_check: string | null
          cloud_provider: 'aws' | 'gcp' | 'azure' | 'oci' | null
          cloud_location: string | null
          deploy_job_execution_name: string | null
          deploy_job_status: 'pending' | 'running' | 'succeeded' | 'failed' | 'unknown' | null
          deploy_job_started_at: string | null
          deploy_job_completed_at: string | null
          deploy_job_error: string | null
          deploy_job_operation: 'install' | 'uninstall' | null
          deploy_job_current_step: number | null
          deploy_job_total_steps: number | null
          deploy_job_step_description: string | null
          root_dns_target: string | null
          root_dns_type: 'cname' | 'a' | null
          root_dns_record_id: string | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          org_id: string
          name?: string | null
          description?: string | null
          installation_key: string
          secret_key?: string | null
          setup_completed?: boolean
          subdomain?: string | null
          reserved_at?: string | null
          deployment_ready?: boolean | null
          last_health_check?: string | null
          cloud_provider?: 'aws' | 'gcp' | 'azure' | 'oci' | null
          cloud_location?: string | null
          deploy_job_execution_name?: string | null
          deploy_job_status?: 'pending' | 'running' | 'succeeded' | 'failed' | 'unknown' | null
          deploy_job_started_at?: string | null
          deploy_job_completed_at?: string | null
          deploy_job_error?: string | null
          deploy_job_operation?: 'install' | 'uninstall' | null
          deploy_job_current_step?: number | null
          deploy_job_total_steps?: number | null
          deploy_job_step_description?: string | null
          root_dns_target?: string | null
          root_dns_type?: 'cname' | 'a' | null
          root_dns_record_id?: string | null
        }
        Update: {
          org_id?: string
          name?: string | null
          description?: string | null
          installation_key?: string
          secret_key?: string | null
          setup_completed?: boolean
          subdomain?: string | null
          reserved_at?: string | null
          deployment_ready?: boolean | null
          last_health_check?: string | null
          cloud_provider?: 'aws' | 'gcp' | 'azure' | 'oci' | null
          cloud_location?: string | null
          deploy_job_execution_name?: string | null
          deploy_job_status?: 'pending' | 'running' | 'succeeded' | 'failed' | 'unknown' | null
          deploy_job_started_at?: string | null
          deploy_job_completed_at?: string | null
          deploy_job_error?: string | null
          deploy_job_operation?: 'install' | 'uninstall' | null
          deploy_job_current_step?: number | null
          deploy_job_total_steps?: number | null
          deploy_job_step_description?: string | null
          root_dns_target?: string | null
          root_dns_type?: 'cname' | 'a' | null
          root_dns_record_id?: string | null
        }
      }
      dns_configurations: {
        Row: {
          id: number
          installation_id: string
          service_name: string
          ip: string
          ssh_record_id: string | null
          route_record_ids: string[]
          route_record_map: Record<string, string> | null
          domain_routes: Array<{ domain: string }>
          created_at: string
          updated_at: string
        }
        Insert: {
          installation_id: string
          service_name: string
          ip: string
          ssh_record_id?: string | null
          route_record_ids?: string[]
          route_record_map?: Record<string, string> | null
          domain_routes?: Array<{ domain: string }>
        }
        Update: {
          installation_id?: string
          service_name?: string
          ip?: string
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
      billing_accounts: {
        Row: {
          id: string
          org_id: string
          stripe_customer_id: string
          stripe_subscription_id: string | null
          billing_status: 'active' | 'past_due' | 'cancelled' | 'trialing' | 'incomplete'
          has_payment_issue: boolean
          current_period_end: string | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          org_id: string
          stripe_customer_id: string
          stripe_subscription_id?: string | null
          billing_status?: 'active' | 'past_due' | 'cancelled' | 'trialing' | 'incomplete'
          has_payment_issue?: boolean
          current_period_end?: string | null
        }
        Update: {
          org_id?: string
          stripe_customer_id?: string
          stripe_subscription_id?: string | null
          billing_status?: 'active' | 'past_due' | 'cancelled' | 'trialing' | 'incomplete'
          has_payment_issue?: boolean
          current_period_end?: string | null
        }
      }
      subscription_items: {
        Row: {
          id: string
          org_id: string
          installation_id: string | null
          stripe_item_id: string
          stripe_price_id: string
          tier: number
          product_name: string
          quantity: number
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          org_id: string
          installation_id?: string | null
          stripe_item_id: string
          stripe_price_id: string
          tier: number
          product_name: string
          quantity?: number
        }
        Update: {
          org_id?: string
          installation_id?: string | null
          stripe_item_id?: string
          stripe_price_id?: string
          tier?: number
          product_name?: string
          quantity?: number
        }
      }
      processed_webhook_events: {
        Row: {
          id: string
          stripe_event_id: string
          event_type: string
          processed_at: string
        }
        Insert: {
          id?: string
          stripe_event_id: string
          event_type: string
          processed_at?: string
        }
        Update: {
          stripe_event_id?: string
          event_type?: string
          processed_at?: string
        }
      }
    }
  }
}
