import { describe, expect, it } from 'vitest'
import { generateUsernameFromEmail } from './username'

describe('generateUsernameFromEmail', () => {
  it('creates a kubernetes-compatible username from email', () => {
    expect(generateUsernameFromEmail('John.Doe_123@example.com')).toBe('john-doe-123')
  })

  it('enforces minimum username length', () => {
    expect(generateUsernameFromEmail('ab@example.com')).toBe('ab-user')
  })
})
