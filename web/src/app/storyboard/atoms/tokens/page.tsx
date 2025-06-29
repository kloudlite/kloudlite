"use client";

import { ComponentShowcase } from "../../_components/component-showcase";
import { Text, Heading } from "@/components/atoms";
import { cn } from "@/lib/utils";
import { Plus, Search, Settings, X } from "lucide-react";

export default function TokensPage() {
  return (
    <div className="space-y-8">
      <div>
        <Heading level={2} className="mb-4">Design Tokens</Heading>
        <Text color="secondary">
          CSS Variables based design tokens that power the Kloudlite design system.
        </Text>
      </div>

      {/* Color Palette */}
      <ComponentShowcase
        title="Color System"
        description="Complete color system using CSS variables for consistent theming"
      >
        <div className="space-y-8">
          {/* Semantic Colors */}
          <div>
            <Text weight="medium" className="mb-4">Semantic Colors</Text>
            <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-6 gap-4">
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-background" />
                <Text size="xs" font="mono">background</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-foreground" />
                <Text size="xs" font="mono">foreground</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-primary" />
                <Text size="xs" font="mono">primary</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-secondary" />
                <Text size="xs" font="mono">secondary</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-muted" />
                <Text size="xs" font="mono">muted</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-accent" />
                <Text size="xs" font="mono">accent</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-destructive" />
                <Text size="xs" font="mono">destructive</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-card" />
                <Text size="xs" font="mono">card</Text>
              </div>
            </div>
          </div>

          {/* Brand Colors */}
          <div>
            <Text weight="medium" className="mb-4">Brand Colors (Blue Primary)</Text>
            <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-10 gap-3">
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-brand-50))]" />
                <Text size="xs" font="mono">brand-50</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-brand-100))]" />
                <Text size="xs" font="mono">brand-100</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-brand-200))]" />
                <Text size="xs" font="mono">brand-200</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-brand-300))]" />
                <Text size="xs" font="mono">brand-300</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-brand-400))]" />
                <Text size="xs" font="mono">brand-400</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-brand-500))]" />
                <Text size="xs" font="mono">brand-500</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-brand-600))]" />
                <Text size="xs" font="mono">brand-600</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-brand-700))]" />
                <Text size="xs" font="mono">brand-700</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-brand-800))]" />
                <Text size="xs" font="mono">brand-800</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-brand-900))]" />
                <Text size="xs" font="mono">brand-900</Text>
              </div>
            </div>
          </div>

          {/* Neutral Colors */}
          <div>
            <Text weight="medium" className="mb-4">Neutral Colors (Gray Scale)</Text>
            <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-11 gap-3">
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-neutral-50))]" />
                <Text size="xs" font="mono">neutral-50</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-neutral-100))]" />
                <Text size="xs" font="mono">neutral-100</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-neutral-200))]" />
                <Text size="xs" font="mono">neutral-200</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-neutral-300))]" />
                <Text size="xs" font="mono">neutral-300</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-neutral-400))]" />
                <Text size="xs" font="mono">neutral-400</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-neutral-500))]" />
                <Text size="xs" font="mono">neutral-500</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-neutral-600))]" />
                <Text size="xs" font="mono">neutral-600</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-neutral-700))]" />
                <Text size="xs" font="mono">neutral-700</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-neutral-800))]" />
                <Text size="xs" font="mono">neutral-800</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-neutral-900))]" />
                <Text size="xs" font="mono">neutral-900</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-neutral-950))]" />
                <Text size="xs" font="mono">neutral-950</Text>
              </div>
            </div>
          </div>

          {/* Status Colors */}
          <div>
            <Text weight="medium" className="mb-4">Status Colors</Text>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              {/* Success */}
              <div>
                <Text size="sm" weight="medium" className="mb-3">Success (Emerald)</Text>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-success-50))]" />
                  <Text size="xs" font="mono">success-50</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-success-500))]" />
                  <Text size="xs" font="mono">success-500</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-success-700))]" />
                  <Text size="xs" font="mono">success-700</Text>
                </div>
              </div>

              {/* Error */}
              <div>
                <Text size="sm" weight="medium" className="mb-3">Error (Rose)</Text>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-error-50))]" />
                  <Text size="xs" font="mono">error-50</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-error-500))]" />
                  <Text size="xs" font="mono">error-500</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-error-700))]" />
                  <Text size="xs" font="mono">error-700</Text>
                </div>
              </div>

              {/* Warning */}
              <div>
                <Text size="sm" weight="medium" className="mb-3">Warning (Amber)</Text>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-warning-50))]" />
                  <Text size="xs" font="mono">warning-50</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-warning-500))]" />
                  <Text size="xs" font="mono">warning-500</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-warning-700))]" />
                  <Text size="xs" font="mono">warning-700</Text>
                </div>
              </div>

              {/* Info */}
              <div>
                <Text size="sm" weight="medium" className="mb-3">Info (Sky)</Text>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-info-50))]" />
                  <Text size="xs" font="mono">info-50</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-info-500))]" />
                  <Text size="xs" font="mono">info-500</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-12 rounded-md border border-border bg-[rgb(var(--color-info-700))]" />
                  <Text size="xs" font="mono">info-700</Text>
                </div>
              </div>
            </div>
          </div>

          {/* Additional Professional Colors */}
          <div className="space-y-8">
            {/* Purple/Indigo */}
            <div>
              <Text weight="medium" className="mb-4">Purple (Indigo) - Highlights & Accents</Text>
              <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-10 gap-3">
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-purple-50))]" />
                  <Text size="xs" font="mono">purple-50</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-purple-100))]" />
                  <Text size="xs" font="mono">purple-100</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-purple-200))]" />
                  <Text size="xs" font="mono">purple-200</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-purple-300))]" />
                  <Text size="xs" font="mono">purple-300</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-purple-400))]" />
                  <Text size="xs" font="mono">purple-400</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-purple-500))]" />
                  <Text size="xs" font="mono">purple-500</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-purple-600))]" />
                  <Text size="xs" font="mono">purple-600</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-purple-700))]" />
                  <Text size="xs" font="mono">purple-700</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-purple-800))]" />
                  <Text size="xs" font="mono">purple-800</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-purple-900))]" />
                  <Text size="xs" font="mono">purple-900</Text>
                </div>
              </div>
            </div>

            {/* Teal */}
            <div>
              <Text weight="medium" className="mb-4">Teal - Positive/Growth Indicators</Text>
              <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-10 gap-3">
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-teal-50))]" />
                  <Text size="xs" font="mono">teal-50</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-teal-100))]" />
                  <Text size="xs" font="mono">teal-100</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-teal-200))]" />
                  <Text size="xs" font="mono">teal-200</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-teal-300))]" />
                  <Text size="xs" font="mono">teal-300</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-teal-400))]" />
                  <Text size="xs" font="mono">teal-400</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-teal-500))]" />
                  <Text size="xs" font="mono">teal-500</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-teal-600))]" />
                  <Text size="xs" font="mono">teal-600</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-teal-700))]" />
                  <Text size="xs" font="mono">teal-700</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-teal-800))]" />
                  <Text size="xs" font="mono">teal-800</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-teal-900))]" />
                  <Text size="xs" font="mono">teal-900</Text>
                </div>
              </div>
            </div>

            {/* Orange */}
            <div>
              <Text weight="medium" className="mb-4">Orange - Notifications/Alerts</Text>
              <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-10 gap-3">
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-orange-50))]" />
                  <Text size="xs" font="mono">orange-50</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-orange-100))]" />
                  <Text size="xs" font="mono">orange-100</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-orange-200))]" />
                  <Text size="xs" font="mono">orange-200</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-orange-300))]" />
                  <Text size="xs" font="mono">orange-300</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-orange-400))]" />
                  <Text size="xs" font="mono">orange-400</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-orange-500))]" />
                  <Text size="xs" font="mono">orange-500</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-orange-600))]" />
                  <Text size="xs" font="mono">orange-600</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-orange-700))]" />
                  <Text size="xs" font="mono">orange-700</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-orange-800))]" />
                  <Text size="xs" font="mono">orange-800</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-orange-900))]" />
                  <Text size="xs" font="mono">orange-900</Text>
                </div>
              </div>
            </div>

            {/* Slate */}
            <div>
              <Text weight="medium" className="mb-4">Slate - Professional Dark Grays</Text>
              <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-11 gap-3">
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-slate-50))]" />
                  <Text size="xs" font="mono">slate-50</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-slate-100))]" />
                  <Text size="xs" font="mono">slate-100</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-slate-200))]" />
                  <Text size="xs" font="mono">slate-200</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-slate-300))]" />
                  <Text size="xs" font="mono">slate-300</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-slate-400))]" />
                  <Text size="xs" font="mono">slate-400</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-slate-500))]" />
                  <Text size="xs" font="mono">slate-500</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-slate-600))]" />
                  <Text size="xs" font="mono">slate-600</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-slate-700))]" />
                  <Text size="xs" font="mono">slate-700</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-slate-800))]" />
                  <Text size="xs" font="mono">slate-800</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-slate-900))]" />
                  <Text size="xs" font="mono">slate-900</Text>
                </div>
                <div className="space-y-2">
                  <div className="h-16 rounded-md border border-border bg-[rgb(var(--color-slate-950))]" />
                  <Text size="xs" font="mono">slate-950</Text>
                </div>
              </div>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      {/* Typography */}
      <ComponentShowcase
        title="Typography Scale"
        description="Font sizes and weights used throughout the design system"
      >
        <div className="space-y-6">
          <div>
            <Text weight="medium" className="mb-4">Font Sizes</Text>
            <div className="space-y-3">
              <div className="flex items-center gap-4">
                <Text size="xs">Extra Small (xs)</Text>
                <Text size="xs" font="mono" color="muted">12px</Text>
              </div>
              <div className="flex items-center gap-4">
                <Text size="sm">Small (sm)</Text>
                <Text size="xs" font="mono" color="muted">14px</Text>
              </div>
              <div className="flex items-center gap-4">
                <Text size="base">Base (base)</Text>
                <Text size="xs" font="mono" color="muted">16px</Text>
              </div>
              <div className="flex items-center gap-4">
                <Text size="lg">Large (lg)</Text>
                <Text size="xs" font="mono" color="muted">18px</Text>
              </div>
              <div className="flex items-center gap-4">
                <Text size="xl">Extra Large (xl)</Text>
                <Text size="xs" font="mono" color="muted">20px</Text>
              </div>
              <div className="flex items-center gap-4">
                <Text size="2xl">2X Large (2xl)</Text>
                <Text size="xs" font="mono" color="muted">24px</Text>
              </div>
            </div>
          </div>

          <div>
            <Text weight="medium" className="mb-4">Font Weights</Text>
            <div className="space-y-3">
              <Text weight="light">Light Weight</Text>
              <Text weight="normal">Normal Weight</Text>
              <Text weight="medium">Medium Weight</Text>
              <Text weight="semibold">Semibold Weight</Text>
              <Text weight="bold">Bold Weight</Text>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      {/* Icon Sizes */}
      <ComponentShowcase
        title="Icon Sizes"
        description="Consistent icon sizing scale"
      >
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <Search className="h-3 w-3" />
            <Text>Extra Small (12px)</Text>
            <Text size="xs" font="mono" color="muted">h-3 w-3</Text>
          </div>
          <div className="flex items-center gap-4">
            <Search className="h-4 w-4" />
            <Text>Small (16px)</Text>
            <Text size="xs" font="mono" color="muted">h-4 w-4</Text>
          </div>
          <div className="flex items-center gap-4">
            <Search className="h-5 w-5" />
            <Text>Base (20px)</Text>
            <Text size="xs" font="mono" color="muted">h-5 w-5</Text>
          </div>
          <div className="flex items-center gap-4">
            <Search className="h-6 w-6" />
            <Text>Large (24px)</Text>
            <Text size="xs" font="mono" color="muted">h-6 w-6</Text>
          </div>
          <div className="flex items-center gap-4">
            <Search className="h-8 w-8" />
            <Text>Extra Large (32px)</Text>
            <Text size="xs" font="mono" color="muted">h-8 w-8</Text>
          </div>
        </div>
      </ComponentShowcase>

      {/* Spacing */}
      <ComponentShowcase
        title="Spacing Scale"
        description="Consistent spacing using Tailwind's scale"
      >
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <div className="h-4 w-1 bg-primary rounded" />
            <Text>1 (4px)</Text>
          </div>
          <div className="flex items-center gap-4">
            <div className="h-4 w-2 bg-primary rounded" />
            <Text>2 (8px)</Text>
          </div>
          <div className="flex items-center gap-4">
            <div className="h-4 w-3 bg-primary rounded" />
            <Text>3 (12px)</Text>
          </div>
          <div className="flex items-center gap-4">
            <div className="h-4 w-4 bg-primary rounded" />
            <Text>4 (16px)</Text>
          </div>
          <div className="flex items-center gap-4">
            <div className="h-4 w-6 bg-primary rounded" />
            <Text>6 (24px)</Text>
          </div>
          <div className="flex items-center gap-4">
            <div className="h-4 w-8 bg-primary rounded" />
            <Text>8 (32px)</Text>
          </div>
        </div>
      </ComponentShowcase>

      {/* Border Radius */}
      <ComponentShowcase
        title="Border Radius"
        description="Consistent border radius scale"
      >
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <div className="space-y-2">
            <div className="h-16 w-16 bg-primary rounded-none" />
            <Text size="xs" font="mono">none</Text>
          </div>
          <div className="space-y-2">
            <div className="h-16 w-16 bg-primary rounded-sm" />
            <Text size="xs" font="mono">sm</Text>
          </div>
          <div className="space-y-2">
            <div className="h-16 w-16 bg-primary rounded" />
            <Text size="xs" font="mono">default</Text>
          </div>
          <div className="space-y-2">
            <div className="h-16 w-16 bg-primary rounded-md" />
            <Text size="xs" font="mono">md</Text>
          </div>
          <div className="space-y-2">
            <div className="h-16 w-16 bg-primary rounded-lg" />
            <Text size="xs" font="mono">lg</Text>
          </div>
          <div className="space-y-2">
            <div className="h-16 w-16 bg-primary rounded-xl" />
            <Text size="xs" font="mono">xl</Text>
          </div>
          <div className="space-y-2">
            <div className="h-16 w-16 bg-primary rounded-2xl" />
            <Text size="xs" font="mono">2xl</Text>
          </div>
          <div className="space-y-2">
            <div className="h-16 w-16 bg-primary rounded-full" />
            <Text size="xs" font="mono">full</Text>
          </div>
        </div>
      </ComponentShowcase>

      {/* Borders */}
      <ComponentShowcase
        title="Borders"
        description="Border styles and colors used in the design system"
      >
        <div className="space-y-6">
          <div>
            <Text weight="medium" className="mb-4">Border Widths</Text>
            <div className="space-y-4">
              <div className="flex items-center gap-4">
                <div className="h-16 w-16 bg-background border-0 border-border rounded" />
                <Text>None (0px)</Text>
                <Text size="xs" font="mono" color="muted">border-0</Text>
              </div>
              <div className="flex items-center gap-4">
                <div className="h-16 w-16 bg-background border border-border rounded" />
                <Text>Default (1px)</Text>
                <Text size="xs" font="mono" color="muted">border</Text>
              </div>
              <div className="flex items-center gap-4">
                <div className="h-16 w-16 bg-background border-2 border-border rounded" />
                <Text>Medium (2px)</Text>
                <Text size="xs" font="mono" color="muted">border-2</Text>
              </div>
              <div className="flex items-center gap-4">
                <div className="h-16 w-16 bg-background border-4 border-border rounded" />
                <Text>Large (4px)</Text>
                <Text size="xs" font="mono" color="muted">border-4</Text>
              </div>
            </div>
          </div>

          <div>
            <Text weight="medium" className="mb-4">Border Colors</Text>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <div className="space-y-2">
                <div className="h-16 w-16 bg-background border-2 border-border rounded" />
                <Text size="xs" font="mono">border</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 w-16 bg-background border-2 border-primary rounded" />
                <Text size="xs" font="mono">primary</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 w-16 bg-background border-2 border-destructive rounded" />
                <Text size="xs" font="mono">destructive</Text>
              </div>
              <div className="space-y-2">
                <div className="h-16 w-16 bg-background border-2 border-muted rounded" />
                <Text size="xs" font="mono">muted</Text>
              </div>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      {/* Shadows */}
      <ComponentShowcase
        title="Shadows"
        description="CSS variable based shadow system for elevation"
      >
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-6">
          <div className="space-y-3">
            <div className="h-16 w-16 bg-card rounded-md shadow-none border border-border" />
            <Text size="xs" font="mono">none</Text>
            <Text size="xs" color="muted">shadow-none</Text>
          </div>
          <div className="space-y-3">
            <div className="h-16 w-16 bg-card rounded-md [box-shadow:var(--shadow-xs)]" />
            <Text size="xs" font="mono">xs</Text>
            <Text size="xs" color="muted">--shadow-xs</Text>
          </div>
          <div className="space-y-3">
            <div className="h-16 w-16 bg-card rounded-md [box-shadow:var(--shadow-sm)]" />
            <Text size="xs" font="mono">sm</Text>
            <Text size="xs" color="muted">--shadow-sm</Text>
          </div>
          <div className="space-y-3">
            <div className="h-16 w-16 bg-card rounded-md [box-shadow:var(--shadow-md)]" />
            <Text size="xs" font="mono">md</Text>
            <Text size="xs" color="muted">--shadow-md</Text>
          </div>
          <div className="space-y-3">
            <div className="h-16 w-16 bg-card rounded-md [box-shadow:var(--shadow-lg)]" />
            <Text size="xs" font="mono">lg</Text>
            <Text size="xs" color="muted">--shadow-lg</Text>
          </div>
          <div className="space-y-3">
            <div className="h-16 w-16 bg-card rounded-md shadow-xl" />
            <Text size="xs" font="mono">xl</Text>
            <Text size="xs" color="muted">shadow-xl</Text>
          </div>
          <div className="space-y-3">
            <div className="h-16 w-16 bg-card rounded-md shadow-2xl" />
            <Text size="xs" font="mono">2xl</Text>
            <Text size="xs" color="muted">shadow-2xl</Text>
          </div>
          <div className="space-y-3">
            <div className="h-16 w-16 bg-card rounded-md shadow-inner border border-border" />
            <Text size="xs" font="mono">inner</Text>
            <Text size="xs" color="muted">shadow-inner</Text>
          </div>
        </div>
      </ComponentShowcase>

      {/* Status Indicators */}
      <ComponentShowcase
        title="Status System"
        description="Status colors and their semantic meanings"
      >
        <div className="space-y-6">
          <div>
            <Text weight="medium" className="mb-4">Status Colors with Meanings</Text>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div className="p-4 rounded-lg border border-[rgb(var(--color-success-200))] bg-[rgb(var(--color-success-50))] dark:border-[rgb(var(--color-success-700))] dark:bg-[rgb(var(--color-success-700)/0.2)]">
                <div className="flex items-center gap-2 mb-2">
                  <div className="h-3 w-3 bg-[rgb(var(--color-success-500))] rounded-full" />
                  <Text weight="medium" color="success">Success / Running</Text>
                </div>
                <Text size="xs" color="success">Active, completed, healthy states</Text>
              </div>
              
              <div className="p-4 rounded-lg border border-[rgb(var(--color-error-200))] bg-[rgb(var(--color-error-50))] dark:border-[rgb(var(--color-error-700))] dark:bg-[rgb(var(--color-error-700)/0.2)]">
                <div className="flex items-center gap-2 mb-2">
                  <div className="h-3 w-3 bg-[rgb(var(--color-error-500))] rounded-full" />
                  <Text weight="medium" color="error">Error / Failed</Text>
                </div>
                <Text size="xs" color="error">Failed, stopped, error states</Text>
              </div>
              
              <div className="p-4 rounded-lg border border-[rgb(var(--color-warning-200))] bg-[rgb(var(--color-warning-50))] dark:border-[rgb(var(--color-warning-700))] dark:bg-[rgb(var(--color-warning-700)/0.2)]">
                <div className="flex items-center gap-2 mb-2">
                  <div className="h-3 w-3 bg-[rgb(var(--color-warning-500))] rounded-full" />
                  <Text weight="medium" color="warning">Warning / Pending</Text>
                </div>
                <Text size="xs" color="warning">Pending, warning, degraded states</Text>
              </div>
              
              <div className="p-4 rounded-lg border border-[rgb(var(--color-info-200))] bg-[rgb(var(--color-info-50))] dark:border-[rgb(var(--color-info-700))] dark:bg-[rgb(var(--color-info-700)/0.2)]">
                <div className="flex items-center gap-2 mb-2">
                  <div className="h-3 w-3 bg-[rgb(var(--color-info-500))] rounded-full" />
                  <Text weight="medium" color="info">Info</Text>
                </div>
                <Text size="xs" color="info">Informational states</Text>
              </div>
              
              <div className="p-4 rounded-lg border border-border bg-muted">
                <div className="flex items-center gap-2 mb-2">
                  <div className="h-3 w-3 bg-muted-foreground rounded-full" />
                  <Text weight="medium">Unknown / Paused</Text>
                </div>
                <Text size="xs" color="muted">Unknown, paused, inactive states</Text>
              </div>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      {/* Breakpoints */}
      <ComponentShowcase
        title="Responsive Breakpoints"
        description="Tailwind CSS breakpoints for responsive design"
      >
        <div className="space-y-4">
          <Text weight="medium">Current breakpoint: 
            <span className="ml-2 px-2 py-1 bg-primary text-primary-foreground rounded text-xs font-mono">
              <span className="sm:hidden">xs (default)</span>
              <span className="hidden sm:inline md:hidden">sm</span>
              <span className="hidden md:inline lg:hidden">md</span>
              <span className="hidden lg:inline xl:hidden">lg</span>
              <span className="hidden xl:inline 2xl:hidden">xl</span>
              <span className="hidden 2xl:inline">2xl</span>
            </span>
          </Text>
          
          <div className="space-y-3">
            <div className="flex items-center justify-between p-3 border border-border rounded">
              <div>
                <Text weight="medium">xs (default)</Text>
                <Text size="xs" color="muted">0px and up</Text>
              </div>
              <Text size="xs" font="mono" color="muted">min-width: 0px</Text>
            </div>
            
            <div className="flex items-center justify-between p-3 border border-border rounded">
              <div>
                <Text weight="medium">sm</Text>
                <Text size="xs" color="muted">640px and up</Text>
              </div>
              <Text size="xs" font="mono" color="muted">min-width: 640px</Text>
            </div>
            
            <div className="flex items-center justify-between p-3 border border-border rounded">
              <div>
                <Text weight="medium">md</Text>
                <Text size="xs" color="muted">768px and up</Text>
              </div>
              <Text size="xs" font="mono" color="muted">min-width: 768px</Text>
            </div>
            
            <div className="flex items-center justify-between p-3 border border-border rounded">
              <div>
                <Text weight="medium">lg</Text>
                <Text size="xs" color="muted">1024px and up</Text>
              </div>
              <Text size="xs" font="mono" color="muted">min-width: 1024px</Text>
            </div>
            
            <div className="flex items-center justify-between p-3 border border-border rounded">
              <div>
                <Text weight="medium">xl</Text>
                <Text size="xs" color="muted">1280px and up</Text>
              </div>
              <Text size="xs" font="mono" color="muted">min-width: 1280px</Text>
            </div>
            
            <div className="flex items-center justify-between p-3 border border-border rounded">
              <div>
                <Text weight="medium">2xl</Text>
                <Text size="xs" color="muted">1536px and up</Text>
              </div>
              <Text size="xs" font="mono" color="muted">min-width: 1536px</Text>
            </div>
          </div>
          
          <div className="p-4 bg-muted rounded-lg">
            <Text size="sm" weight="medium" className="mb-2">Responsive Example</Text>
            <div className="h-8 rounded bg-primary">
              <div className="h-full bg-primary/80 rounded sm:bg-[rgb(var(--color-success-500))] md:bg-[rgb(var(--color-warning-500))] lg:bg-[rgb(var(--color-error-500))] xl:bg-[rgb(var(--color-info-500))] 2xl:bg-primary transition-colors" />
            </div>
            <Text size="xs" color="muted" className="mt-2">
              This bar changes color at different breakpoints
            </Text>
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}