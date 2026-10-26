// design-sync only. @kloudlite/ui imports `cn` from @kloudlite/lib, whose
// barrel evaluates env.ts (process.env.NEXT_PUBLIC_*) at module load. That is
// fine under Next.js but throws in a bare browser, which is where preview
// cards and every design built with this DS run. This module is bundled ahead
// of the DS entry, so `process` exists by the time env.ts evaluates.
;(globalThis as unknown as { process?: unknown }).process ??= {
  env: { NODE_ENV: 'development' },
}
export {}
