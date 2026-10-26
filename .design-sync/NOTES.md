# design-sync notes — @kloudlite/ui

Repo-specific gotchas for future syncs. Read this before re-running the driver.

## Setup quirks

- **The package has no build step.** `package.json` `main`/`types` point at `src/index.ts`,
  so the converter runs in synth-entry mode and reads `src/` directly. There is no `dist/`
  to build and no `--entry` to pass. `cfg.tsconfig` is set so esbuild resolves the `@/*` and
  `@kloudlite/lib` path aliases.
- **A workspace self-symlink is required.** The converter resolves the package at
  `<node-modules>/<pkg>`, but bun does not self-link a workspace package. Recreate it on a
  fresh clone (it is inside gitignored `node_modules`, so it never survives one):

      ln -sfn ../.. web/packages/ui/node_modules/@kloudlite/ui

  Without it the build dies with `ENOENT … @kloudlite/ui/package.json`.
- Use `--node-modules web/packages/ui/node_modules` (not the repo root). The package has its
  own `react`, `react-dom`, and `@types/react`; the root does not.

## `process is not defined` — solved, do not regress

`@kloudlite/ui` imports `cn` from `@kloudlite/lib`, and that barrel evaluates
`src/env.ts` (`export const env = validateEnv()`) at module load, reading
`process.env.NEXT_PUBLIC_*`. Under Next.js that is fine; in a bare browser it throws, which
killed **every** preview on the first run.

Fix: `web/packages/ui/.ds-shim.ts`, wired through `cfg.extraEntries`. `extraEntries` modules
are emitted ahead of the DS entry in `.bundle-entry.mjs`, so the shim defines `globalThis.process`
before `env.ts` evaluates. Keep the shim as the FIRST `extraEntries` element.

The real upstream fix would be for `@kloudlite/ui` to import `cn` from a leaf module rather
than the `@kloudlite/lib` barrel — that would drop server-only env validation out of a
browser bundle entirely. Worth doing; until then the shim is load-bearing.

## CSS: Tailwind v4 must be compiled first

`src/globals.css` is Tailwind v4 **source** (`@import 'tailwindcss'`), not a stylesheet.
`cfg.buildCmd` compiles it to `.ds-css/ui.css`, which is what `cfg.cssEntry` points at.
**Always run `cfg.buildCmd` before `package-build.mjs`** — the driver does not run it for you.

- `.ds-css/entry.css` (committed) is the compile input: it pulls in `globals.css`, `@source`s
  the component sources, the authored previews, and `safelist.txt`, and binds `--font-sans` /
  `--font-mono`, which the DS references but never defines (the console app supplies them via
  `next/font`).
- `.ds-css/ui.css` is generated output and gitignored.
- **`.ds-css/safelist.txt` exists because of a fan-out race.** Tailwind only emits utilities it
  finds in scanned files, so a class used for the first time in a new preview is missing from
  the compiled CSS until a recompile — and subagents may not recompile (it is a shared file).
  The safelist pre-emits ~950 common utilities so authored previews have a guaranteed
  vocabulary. **Previews should stick to safelisted classes.** If a preview needs something
  outside it, add the class to `safelist.txt` rather than relying on the source scan.
- Fonts load remotely via a Google Fonts `@import` in `entry.css` → `[FONT_REMOTE]`,
  informational, no action. Nothing ships in `fonts/` by design.

## Grouping

The DS is a flat `src/*.tsx` tree, so everything landed in one `general` group by default.
`web/packages/ui/.ds-docs/<Name>.md` holds a frontmatter-only stub (`category: <Group>`) per
component, bound via `cfg.docsDir`. A frontmatter-only stub sets the group **without**
replacing the synthesized `.prompt.md` (the converter only substitutes when the doc has a
body), so these are pure grouping metadata. They are committed and are the natural place to
add real prose docs later — a body added to one of these files becomes that component's
`.prompt.md`.

`cfg.componentSrcMap` maps all 285 exports to their defining source file. It is not sparse
because it cannot be: the compound parts (`CardHeader`, `SelectTrigger`, …) live in their
parent's file, so the converter's `<Name>.tsx` fuzzy-find missed 233 of them and they fell
back to generic prop bodies. With the map, `.d.ts` extraction pulls real Radix props and
JSDoc (e.g. `AccordionItem.value: string`). Regenerate it if components are added.

## Known render warns

- `[DTS_STYLE_SYSTEM] filtering @types/react props` — expected. React's CSS-shorthand-named
  props are filtered from prop bodies; none of this DS's real API is affected.
- Editor/TypeScript errors inside `.design-sync/previews/*.tsx` ("Cannot find module 'react'",
  "no interface JSX.IntrinsicElements") are **noise** — that directory is in no tsconfig.
  The converter compiles previews with esbuild plus the package tsconfig paths. Ignore them.

## Re-sync risks

- **The self-symlink and the playwright install do not survive a fresh clone.** Both are in
  gitignored territory. Recreate the symlink (above); reinstall browsers with
  `cd .ds-sync && npm i playwright && npx playwright install chromium`.
- **`.ds-css/ui.css` is generated and gitignored** — a fresh clone must run `cfg.buildCmd`
  before the first build or `cfg.cssEntry` will not resolve.
- **`cfg.componentSrcMap` and `.ds-docs/` stubs enumerate all 285 exports.** Adding or
  renaming a component in `web/packages/ui/src` leaves them stale: the new component gets a
  generic prop body and lands in the `general` group. Regenerate both when the component set
  changes.
- **The `process` shim is tied to upstream code.** If `@kloudlite/lib`'s barrel stops
  evaluating `env.ts` at import time, the shim becomes dead weight and can be dropped; if
  more server-only modules join that barrel, the shim may need more than `env`.
- Fonts are fetched from Google Fonts at render time. An offline or CSP-restricted render
  environment falls back to system fonts.

## Upstream source bugs found while authoring previews

These are real defects in `web/packages/ui/src`, found by rendering every component.
They are NOT design-sync problems and were deliberately left unfixed (the sync does not
touch DS source). Worth a separate PR:

- **`kloudlite-logo.tsx` — the logo cannot be resized.** Its `<svg>` has a hardcoded
  `height="22"`, and the class `h-5.5` is not a real Tailwind class so it is dropped.
  `className` only reaches the outer wrapper, never the `svg`. Previews work around it with
  an inline `style` zoom. Fix: drop the hardcoded `height` attribute and forward `className`
  to the `svg`.
- **`input-otp.tsx` — mangled class names.** `InputOTPSlot` carries
  `first:rounded-none-md` / `last:rounded-none-md`, which look like a botched find/replace
  (probably `rounded-l-md` / `rounded-r-md` originally). They are inert, so every slot
  renders square.
- **`collapsible.tsx` is a bare Radix re-export with no DS styling at all** — every visual
  comes from the consumer. Fine if intentional; surprising if not.

## Preview authoring conventions (learned during the fan-out)

- **Stick to `safelist.txt` classes in previews.** The two calibration files (`Card.tsx`,
  `Select.tsx`) originally used arbitrary values (`w-[380px]`, `w-[260px]`); they rendered
  only because a recompile happened to pick them up, and they misled the fan-out agents.
  They now use `w-96` / `w-64`. Keep reference previews safelist-clean.
- **The safelist is not the only source of compiled classes.** `.ds-css/entry.css` also
  `@source`s the DS sources and the previews, so anything used in `web/packages/ui/src`
  compiles whether or not it is safelisted. A fan-out agent reported
  `data-[panel-group-direction=vertical]:*` as "missing" after grepping only `safelist.txt`;
  those classes are in fact present in `ui.css`. When checking whether a class will render,
  grep the compiled `web/packages/ui/.ds-css/ui.css`, not the safelist.
- **`react-resizable-panels@4` treats numeric `defaultSize`/`minSize`/`maxSize` as PIXELS**,
  despite the emitted `.d.ts` typing them `string | number`. `defaultSize={45}` yields a 45px
  pane. Use percent strings (`"45%"`). All resizable previews do.
- **Radix `*Sub` submenus need care under the capture harness's StrictMode.** `defaultOpen`
  alone gets reset by the effect double-invoke (`onOpenChange(false)` on cleanup). Use a
  controlled `open`, and for Menubar additionally `forceMount` on the `*SubContent`.
- **Charts must set `isAnimationActive={false}`** on every recharts series, or the screenshot
  catches a mid-animation frame and truncates the data.
- **Remote images never load in the render check.** Use inline
  `data:image/svg+xml;utf8,` + `encodeURIComponent` (that is how `AvatarImage` gets a real
  image-loaded cell).
- **Pin any date-dependent preview** (`Calendar` uses `defaultMonth`), or the screenshot
  drifts with the wall clock and every re-sync looks like a visual change.
- **Put an open overlay at the root of its exported cell.** Declaring it inside a decorated
  wrapper composites the portal scrim over the wrapper and washes the card out.

## Confirmed Tailwind v3 → v4 migration bug in `sidebar.tsx`

`w-[--sidebar-width]` compiles to literal `width: --sidebar-width` — **invalid CSS**, which
every browser drops. Verified in the compiled `ui.css`:

    .w-\[--sidebar-width\] { width: --sidebar-width; }

Tailwind v3 auto-wrapped a bare custom property in `var()`; **v4 does not**. The v4 spelling
is `w-(--sidebar-width)` (or `w-[var(--sidebar-width)]`). Same defect on
`w-[--sidebar-width-icon]`. Net effect: the sidebar has no width from these classes at all —
it sizes to content. This is not a preview artifact; it affects the console app too.

Other `rounded-none-*` / `rounded-none[...]` malformed classes recorded above are from the
same square-corners migration and are similarly inert.

## Investigated and dismissed: "`asChild` is broken bundle-wide"

A fan-out agent reported that `asChild` on `ContextMenuTrigger` drops every Radix prop and
hypothesised a duplicate-React / Slot identity problem between `_ds_bundle.js` and
`_vendor/react.js`, "likely bundle-wide". **Both parts checked out false:**

- `_ds_bundle.js` references `window.React` and contains **zero** inlined React internals
  (no `react.element` / `react.transitional.element` symbols). There is exactly one React.
- **78 authored previews use `asChild` and every one graded good**, including
  `AlertDialogTrigger` rendering a styled `Button` through it. A bundle-wide Slot failure
  would have shown as unstyled triggers in all of them.

Whatever the agent hit is specific to `ContextMenuTrigger`. Do not go chasing a duplicate-React
bug on a re-sync — this is the record that it was already looked for and is not there.

## Statically unrenderable, by component (skipped deliberately)

Animations and transitions everywhere; plus: drag-to-resize (Resizable), drag-to-dismiss and
snap points (Drawer), hover-to-open (Tooltip/HoverCard/NavigationMenu, submenu-on-hover),
keyboard navigation and item hover highlight (all menus), scroll position (ScrollArea),
`ThemeSwitcher`'s dropdown (click-only, no `open` prop), `Sidebar`'s icon-collapsed and mobile
Sheet variants, `SidebarMenuAction showOnHover` (`md:opacity-0` at the 900px capture width),
`Field orientation="responsive"` (container query never reaches `@md` in a cell), and
indeterminate `Progress` (the DS translates by `100 - (value || 0)`, so it is pixel-identical
to `value={0}`).
- **Charts: `recharts` is not re-exported from `@kloudlite/ui`.** Chart previews import it
  directly; it resolves fine with `--node-modules web/packages/ui/node_modules`. Chart config
  colours must be `'var(--chart-1)'` — **not** `hsl(var(--chart-1))`; the DS tokens are bare
  `oklch()` values, not HSL triples.
- **`sonner`'s `toast` is likewise not re-exported.** A preview importing `sonner` directly
  inlined a *second* sonner with its own toast store, so toasts never reached the mounted
  `Toaster`. Fixed with `cfg.extraEntries: ["sonner"]` + `cfg.storyImports.shim: ["sonner"]`,
  routing preview imports to the bundle's copy. The cleaner upstream fix is re-exporting
  `toast` from `src/sonner.tsx`.
- **`ContextMenu` (Root) exposes no `open`/`defaultOpen`** — Radix gives it only
  `{children, onOpenChange, dir, modal}`. Its previews dispatch a real `contextmenu`
  MouseEvent on mount to open it. Nothing else in the DS needs that trick.
- **`ChartTooltipContent` uses `min-w-[8rem]`**; long config labels wrap awkwardly.

## Ordering lesson — do not repeat

**Apply `[GRID_OVERFLOW]` remedies BEFORE the authoring/grading fan-out.**
`cfg.overrides.<Name>.cardMode` / `primaryStory` are part of the preview contract, so adding
them to 233 components after grading cleared all 828 grades ("0 carried forward") and forced
a complete re-grade. Sequence it as: build → validate once → apply grid-overflow overrides →
then fan out to author and grade.
