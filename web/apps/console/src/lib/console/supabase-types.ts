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
          aca_job_execution_name: string | null
          aca_job_status: 'pending' | 'running' | 'succeeded' | 'failed' | 'unknown' | null
          aca_job_started_at: string | null
          aca_job_completed_at: string | null
          aca_job_error: string | null
          aca_job_operation: 'install' | 'uninstall' | null
          aca_job_current_step: number | null
          aca_job_total_steps: number | null
          aca_job_step_description: string | null
          root_dns_target: string | null
          root_dns_type: 'cname' | 'a' | null
          root_dns_record_id: string | null
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
          aca_job_execution_name?: string | null
          aca_job_status?: 'pending' | 'running' | 'succeeded' | 'failed' | 'unknown' | null
          aca_job_started_at?: string | null
          aca_job_completed_at?: string | null
          aca_job_error?: string | null
          aca_job_operation?: 'install' | 'uninstall' | null
          aca_job_current_step?: number | null
          aca_job_total_steps?: number | null
          aca_job_step_description?: string | null
          root_dns_target?: string | null
          root_dns_type?: 'cname' | 'a' | null
          root_dns_record_id?: string | null
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
          aca_job_execution_name?: string | null
          aca_job_status?: 'pending' | 'running' | 'succeeded' | 'failed' | 'unknown' | null
          aca_job_started_at?: string | null
          aca_job_completed_at?: string | null
          aca_job_error?: string | null
          aca_job_operation?: 'install' | 'uninstall' | null
          aca_job_current_step?: number | null
          aca_job_total_steps?: number | null
          aca_job_step_description?: string | null
          root_dns_target?: string | null
          root_dns_type?: 'cname' | 'a' | null
          root_dns_record_id?: string | null
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
      subscription_plans: {
        Row: {
          id: string
          razorpay_plan_id: string | null
          tier: number
          name: string
          amount_per_user: number
          base_fee: number
          currency: string
          monthly_hours: number
          overage_rate: number
          cpu: number
          ram: string
          storage: string
          auto_suspend: string
          description: string | null
          annual_discount_pct: number
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          razorpay_plan_id?: string | null
          tier: number
          name: string
          amount_per_user: number
          base_fee?: number
          currency?: string
          monthly_hours?: number
          overage_rate?: number
          cpu: number
          ram: string
          storage: string
          auto_suspend: string
          description?: string | null
          annual_discount_pct?: number
        }
        Update: {
          razorpay_plan_id?: string | null
          tier?: number
          name?: string
          amount_per_user?: number
          base_fee?: number
          currency?: string
          monthly_hours?: number
          overage_rate?: number
          cpu?: number
          ram?: string
          storage?: string
          auto_suspend?: string
          description?: string | null
          annual_discount_pct?: number
        }
      }
      subscriptions: {
        Row: {
          id: string
          installation_id: string
          plan_id: string
          razorpay_subscription_id: string | null
          razorpay_customer_id: string | null
          status:
            | 'created'
            | 'authenticated'
            | 'active'
            | 'paused'
            | 'cancelled'
            | 'expired'
          quantity: number
          billing_period: 'monthly' | 'annual'
          current_start: string | null
          current_end: string | null
          scheduled_billing_period: 'monthly' | 'annual' | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          installation_id: string
          plan_id: string
          razorpay_subscription_id?: string | null
          razorpay_customer_id?: string | null
          status?:
            | 'created'
            | 'authenticated'
            | 'active'
            | 'paused'
            | 'cancelled'
            | 'expired'
          quantity?: number
          billing_period?: 'monthly' | 'annual'
          current_start?: string | null
          current_end?: string | null
          scheduled_billing_period?: 'monthly' | 'annual' | null
        }
        Update: {
          plan_id?: string
          razorpay_subscription_id?: string | null
          razorpay_customer_id?: string | null
          status?:
            | 'created'
            | 'authenticated'
            | 'active'
            | 'paused'
            | 'cancelled'
            | 'expired'
          quantity?: number
          billing_period?: 'monthly' | 'annual'
          current_start?: string | null
          current_end?: string | null
          scheduled_billing_period?: 'monthly' | 'annual' | null
        }
      }
      invoices: {
        Row: {
          id: string
          subscription_id: string
          installation_id: string
          razorpay_invoice_id: string | null
          razorpay_payment_id: string | null
          amount: number
          currency: string
          status: 'issued' | 'paid' | 'expired' | 'cancelled'
          billing_start: string | null
          billing_end: string | null
          paid_at: string | null
          created_at: string
        }
        Insert: {
          id?: string
          subscription_id: string
          installation_id: string
          razorpay_invoice_id?: string | null
          razorpay_payment_id?: string | null
          amount: number
          currency?: string
          status?: 'issued' | 'paid' | 'expired' | 'cancelled'
          billing_start?: string | null
          billing_end?: string | null
          paid_at?: string | null
        }
        Update: {
          razorpay_invoice_id?: string | null
          razorpay_payment_id?: string | null
          amount?: number
          currency?: string
          status?: 'issued' | 'paid' | 'expired' | 'cancelled'
          billing_start?: string | null
          billing_end?: string | null
          paid_at?: string | null
        }
      }
      renewal_jobs: {
        Row: {
          id: string
          installation_id: string
          job_type: 'renewal' | 'expire'
          scheduled_at: string
          status: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled'
          attempts: number
          last_error: string | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          installation_id: string
          job_type: 'renewal' | 'expire'
          scheduled_at: string
          status?: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled'
          attempts?: number
          last_error?: string | null
        }
        Update: {
          job_type?: 'renewal' | 'expire'
          scheduled_at?: string
          status?: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled'
          attempts?: number
          last_error?: string | null
        }
      }
      processed_webhook_events: {
        Row: {
          id: string
          razorpay_event_id: string
          event_type: string
          processed_at: string
        }
        Insert: {
          id?: string
          razorpay_event_id: string
          event_type: string
          processed_at?: string
        }
        Update: {
          razorpay_event_id?: string
          event_type?: string
          processed_at?: string
        }
      }
    }
  }
}
