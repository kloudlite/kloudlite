import { describe, expect, it } from 'vitest'
import { generateUsernameFromEmail } from './username'

describe('generateUsernameFromEmail', () => {
  it('normalizes invalid symbols and case', () => {
    expect(generateUsernameFromEmail('Test_User+Dev@example.com')).toBe('test-userdev')
  })

  it('trims to maximum supported length', () => {
    const long = `${'a'.repeat(80)}@example.com`
    expect(generateUsernameFromEmail(long)).toHaveLength(63)
  })
})
