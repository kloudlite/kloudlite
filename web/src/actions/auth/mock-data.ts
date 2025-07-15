import { User } from '@/lib/auth/types'

// Mock user database
export const mockUsers: (User & { password: string })[] = [
  {
    id: '1',
    email: 'user@example.com',
    password: 'password123', // In real app, this would be hashed
    name: 'Test User',
    verified: true,
    createdAt: new Date('2024-01-01'),
    updatedAt: new Date('2024-01-01'),
  },
  {
    id: '2',
    email: 'admin@kloudlite.io',
    password: 'admin123',
    name: 'Admin User',
    verified: true,
    createdAt: new Date('2024-01-01'),
    updatedAt: new Date('2024-01-01'),
  },
]

// Mock sessions (in-memory storage)
export const mockSessions = new Map<string, { userId: string; expiresAt: Date }>()

// Mock password reset tokens
export const mockResetTokens = new Map<string, { email: string; expiresAt: Date }>()

// Mock verification tokens
export const mockVerificationTokens = new Map<string, { email: string; expiresAt: Date }>()

// Helper to generate mock tokens
export function generateMockToken(): string {
  return Math.random().toString(36).substring(2) + Date.now().toString(36)
}

// Helper to create session
export function createMockSession(userId: string, rememberMe: boolean = false): string {
  const sessionId = generateMockToken()
  const expiresAt = new Date()
  expiresAt.setHours(expiresAt.getHours() + (rememberMe ? 24 * 30 : 24)) // 30 days or 1 day
  
  mockSessions.set(sessionId, { userId, expiresAt })
  return sessionId
}

// Helper to validate session
export function validateMockSession(sessionId: string): User | null {
  const session = mockSessions.get(sessionId)
  if (!session) return null
  
  if (new Date() > session.expiresAt) {
    mockSessions.delete(sessionId)
    return null
  }
  
  const user = mockUsers.find(u => u.id === session.userId)
  if (!user) return null
  
  const { password, ...userWithoutPassword } = user
  return userWithoutPassword
}