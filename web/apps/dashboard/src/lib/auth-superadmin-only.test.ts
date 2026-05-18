import { describe, expect, test } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const authSource = readFileSync(join(process.cwd(), 'src/lib/auth.ts'), 'utf8')

describe('dashboard auth provider policy', () => {
  test('only exposes the super-admin token credential', () => {
    expect(authSource).toContain('superadminToken')
    expect(authSource).not.toContain("email: { label: 'Email'")
    expect(authSource).not.toContain("password: { label: 'Password'")
  })

  test('does not register OAuth providers', () => {
    expect(authSource).not.toContain("next-auth/providers/google")
    expect(authSource).not.toContain("next-auth/providers/github")
    expect(authSource).not.toContain("next-auth/providers/microsoft-entra-id")
  })
})
