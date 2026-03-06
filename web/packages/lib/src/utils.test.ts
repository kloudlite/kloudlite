import { describe, expect, it } from 'vitest'
import {
  cn,
  formatWorkspaceName,
  formatResourceName,
  getResourceName,
  getResourceOwner,
} from './utils'

describe('utils', () => {
  it('merges class names with tailwind conflict resolution', () => {
    expect(cn('px-2 py-2', 'px-4')).toBe('py-2 px-4')
  })

  it('formats workspace display names', () => {
    expect(formatWorkspaceName('alice', 'api')).toBe('alice/api')
  })

  it('formats namespaced resource labels', () => {
    expect(formatResourceName('alice--checkout')).toBe('alice/checkout')
    expect(formatResourceName('checkout')).toBe('checkout')
  })

  it('extracts resource name and owner', () => {
    expect(getResourceName('alice--checkout')).toBe('checkout')
    expect(getResourceName('checkout')).toBe('checkout')
    expect(getResourceOwner('alice--checkout')).toBe('alice')
    expect(getResourceOwner('checkout')).toBeNull()
  })
})
