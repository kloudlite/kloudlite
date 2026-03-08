/**
 * Supabase PII Database Type Definitions
 * Separate database for personally identifiable information
 */

export type PiiDatabase = {
  public: {
    Tables: {
      users: {
        Row: {
          user_id: string
          email: string
          name: string
          providers: ('github' | 'google' | 'azure-ad' | 'email')[]
          created_at: string
          updated_at: string
        }
        Insert: {
          user_id: string
          email: string
          name: string
          providers?: ('github' | 'google' | 'azure-ad' | 'email')[]
        }
        Update: {
          email?: string
          name?: string
          providers?: ('github' | 'google' | 'azure-ad' | 'email')[]
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
      contact_messages: {
        Row: {
          id: string
          name: string
          email: string
          subject: string
          message: string
          ip_address: string | null
          user_agent: string | null
          created_at: string
          read_at: string | null
          replied_at: string | null
          status: 'new' | 'read' | 'replied' | 'archived'
        }
        Insert: {
          id?: string
          name: string
          email: string
          subject: string
          message: string
          ip_address?: string | null
          user_agent?: string | null
          status?: 'new' | 'read' | 'replied' | 'archived'
        }
        Update: {
          status?: 'new' | 'read' | 'replied' | 'archived'
          read_at?: string | null
          replied_at?: string | null
        }
      }
    }
  }
}
