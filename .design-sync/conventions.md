## How to build with Kloudlite UI

This is a shadcn/Radix-style library styled with Tailwind v4 semantic tokens. Compose the
exported components and use the token utilities below for your own layout — do not invent a
parallel colour or spacing vocabulary.

### Setup

No global provider is required. Link `styles.css` and load `_ds_bundle.js`; components work
immediately. Three exceptions, all local rather than app-wide:

- **Tooltips** must have a `TooltipProvider` ancestor (wrap the tooltip, or the screen).
- **Sidebar** parts must be inside `SidebarProvider`. Inside a fixed-height container use
  `<Sidebar collapsible="none">` — the default is `fixed inset-y-0 h-svh` and covers the page.
- **Form** parts (`FormField`, `FormItem`, `FormLabel`, `FormControl`, `FormMessage`) require
  `react-hook-form`: call `useForm()` and wrap in `<Form {...form}>`. For a plain layout with
  no form library, use `Field` / `FieldLabel` / `FieldDescription` / `FieldError` instead.

**Dark mode** is a `dark` class on an ancestor (`<div className="dark">` or `<html>`); every
token has a dark value. `Toaster` must be mounted once for `toast()` to appear.

### Styling idiom — use these token utilities

Colours are **semantic pairs**: a surface and its matching foreground. Never hardcode hex,
never use raw Tailwind palette colours like `bg-blue-500`.

| Family | Utilities |
|---|---|
| Page | `bg-background` `text-foreground` |
| Surfaces | `bg-card`/`text-card-foreground`, `bg-popover`/`text-popover-foreground`, `bg-sidebar`/`text-sidebar-foreground` |
| Actions | `bg-primary`/`text-primary-foreground`, `bg-secondary`/`text-secondary-foreground`, `bg-accent`/`text-accent-foreground` |
| Status | `bg-destructive` `bg-success` `bg-warning` `bg-info` (each with `-foreground`) |
| De-emphasis | `bg-muted` `text-muted-foreground` |
| Lines | `border-border` `border-input` `ring-ring` |
| Charts | `bg-chart-1` … `bg-chart-5` (in chart configs use `var(--chart-1)`, not `hsl(...)`) |
| Brand OAuth | `bg-github` `bg-google` `bg-microsoft` (each with `-foreground`) |
| Type | `font-sans` (Open Sans), `font-mono` (IBM Plex Mono) |

Everything else is stock Tailwind v4: `flex` `grid` `gap-4` `p-6` `w-96` `text-sm`
`font-medium` `sm:` `md:` `hover:` `dark:`.

**Corners are square.** `--radius` is `0`, so `rounded-md` / `rounded-lg` resolve to no
radius. That flat, sharp look is deliberate — do not add `rounded-*` to "soften" components.

### Where the truth is

- `styles.css` and its `@import` closure — every token value and component style.
- `components/<group>/<Name>/<Name>.prompt.md` — real usage examples per component.
- `components/<group>/<Name>/<Name>.d.ts` — the exact props, including Radix ones.

Read the component's `.prompt.md` before using it; the examples are real rendered code.

### Idiomatic example

```jsx
const { Card, CardHeader, CardTitle, CardDescription, CardContent, Badge, Button } = window.KloudliteUI

<Card className="w-96">
  <CardHeader>
    <div className="flex items-start justify-between gap-3">
      <div className="space-y-1.5">
        <CardTitle>kloudlite-dev</CardTitle>
        <CardDescription>Installation · Kloudlite Cloud</CardDescription>
      </div>
      <Badge>Active</Badge>
    </div>
  </CardHeader>
  <CardContent className="space-y-3 text-sm">
    <div className="flex justify-between">
      <span className="text-muted-foreground">Region</span>
      <span className="font-medium">ap-south-1</span>
    </div>
    <Button size="sm" className="mt-2">Open workspace</Button>
  </CardContent>
</Card>
```

Note the pattern: library components carry the chrome; your own glue is plain Tailwind using
the semantic tokens (`text-muted-foreground`, `font-medium`, `gap-3`).
