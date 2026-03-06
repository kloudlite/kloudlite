import { describe, expect, it } from 'vitest'
import { generateUsernameFromEmail } from './username'

describe('generateUsernameFromEmail', () => {
  it('normalizes case and replaces dot/underscore with hyphen', () => {
    expect(generateUsernameFromEmail('John.Doe_123@example.com')).toBe('john-doe-123')
  })

  it('removes unsupported characters from local part', () => {
    expect(generateUsernameFromEmail('Test_User+Dev@example.com')).toBe('test-userdev')
  })

  it('pads usernames shorter than 3 characters', () => {
    expect(generateUsernameFromEmail('ab@example.com')).toBe('ab-user')
  })

  it('truncates usernames to 63 chars and trims trailing hyphens after truncation', () => {
    const longLocalPart = `${'a'.repeat(62)}-extra`
    const username = generateUsernameFromEmail(`${longLocalPart}@example.com`)
    expect(username.length).toBeLessThanOrEqual(63)
    expect(username.endsWith('-')).toBe(false)
  })

  it('returns empty string for empty input', () => {
    expect(generateUsernameFromEmail('')).toBe('')
  })
})
