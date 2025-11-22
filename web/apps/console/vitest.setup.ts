import '@testing-library/jest-dom'

// Mock environment variables
process.env.SUPABASE_URL = 'https://test.supabase.co'
process.env.SUPABASE_KEY = 'test-key'
process.env.CLOUDFLARE_ACCOUNT_ID = 'test-account-id'
process.env.CLOUDFLARE_API_TOKEN = 'test-token'
process.env.CLOUDFLARE_ZONE_ID = 'test-zone-id'
process.env.CLOUDFLARE_DNS_DOMAIN = 'test.dev'
process.env.NEXTAUTH_SECRET = 'test-secret-key-with-at-least-32-characters'
process.env.NEXTAUTH_URL = 'http://localhost:3000'
