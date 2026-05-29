import { describe, expect, test } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const srcRoot = join(process.cwd(), 'src')

describe('admin header sizing', () => {
  test('brand label and profile trigger use the same text size as nav items', () => {
    const layout = readFileSync(join(srcRoot, 'app/admin/layout.tsx'), 'utf8')
    const profileDropdown = readFileSync(join(srcRoot, 'app/admin/_components/admin-profile-dropdown.tsx'), 'utf8')

    expect(layout).toContain('text-muted-foreground text-sm font-medium')
    expect(layout).not.toContain('text-muted-foreground text-lg font-medium')
    expect(profileDropdown).toContain('className="gap-1 text-sm"')
  })
})
