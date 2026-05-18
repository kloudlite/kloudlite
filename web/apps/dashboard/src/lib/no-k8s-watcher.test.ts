import { describe, expect, test } from 'vitest'
import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

const root = process.cwd()

function read(path: string) {
  return readFileSync(join(root, path), 'utf8')
}

function files(dir: string): string[] {
  const absolute = join(root, dir)
  if (!existsSync(absolute)) return []
  return readdirSync(absolute).flatMap((entry) => {
    const entryPath = join(absolute, entry)
    if (statSync(entryPath).isDirectory()) return files(relative(root, entryPath))
    return [relative(root, entryPath)]
  })
}

describe('dashboard no longer starts Kubernetes watchers', () => {
  test('server instrumentation does not initialize Kubernetes watches', () => {
    const instrumentation = read('src/instrumentation.ts')

    expect(instrumentation).not.toContain('initializeWatchers')
    expect(instrumentation).not.toContain('k8s-watcher')
    expect(instrumentation).not.toContain('K8s watchers')
  })

  test('dashboard source does not contain active K8S watcher implementation', () => {
    const source = files('src').filter((path) => /\.(ts|tsx)$/.test(path) && !/\.test\.(ts|tsx)$/.test(path))
    const offenders = source.flatMap((path) => {
      const content = read(path)
      return ['K8S-WATCHER', 'createWatch()', 'list-then-watch']
        .filter((pattern) => content.includes(pattern))
        .map((pattern) => `${path}: ${pattern}`)
    })

    expect(offenders).toEqual([])
  })

  test('admin user and machine type actions use platform-api instead of Kubernetes repositories', () => {
    const actionFiles = [
      'src/app/actions/user.actions.ts',
      'src/app/actions/machine-type.actions.ts',
      'src/app/actions/work-machine.actions.ts',
    ]

    const offenders = actionFiles.flatMap((path) => {
      const content = read(path)
      return ['@kloudlite/lib/k8s', 'resourceStore', 'userRepository', 'machineTypeRepository']
        .filter((pattern) => content.includes(pattern))
        .map((pattern) => `${path}: ${pattern}`)
    })

    expect(offenders).toEqual([])
  })
})
