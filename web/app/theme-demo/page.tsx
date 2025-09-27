import {
  ThemeToggleClassic,
  ThemeToggleDropdown,
  ThemeToggleCompact,
  ThemeTogglePill,
  ThemeToggleSelect,
  ThemeToggleRadio,
  ThemeToggleText,
  ThemeToggleAnimated,
  ThemeToggleTooltip,
  ThemeTogglePalette,
} from "@/components/theme-toggle-variants"

export default function ThemeDemoPage() {
  return (
    <div className="container mx-auto p-8 space-y-8">
      <h1 className="text-3xl font-bold mb-8">Theme Switcher Variants</h1>
      
      <div className="grid gap-8 md:grid-cols-2">
        {/* Variant 1 */}
        <div className="border rounded-lg p-6 space-y-4">
          <div>
            <h3 className="font-semibold mb-2">1. Classic Toggle</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Simple sun/moon icon that toggles between light and dark
            </p>
          </div>
          <div className="flex items-center justify-center p-4 bg-muted/30 rounded">
            <ThemeToggleClassic />
          </div>
        </div>

        {/* Variant 2 */}
        <div className="border rounded-lg p-6 space-y-4">
          <div>
            <h3 className="font-semibold mb-2">2. Dropdown Menu</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Click to show options for light, dark, and system
            </p>
          </div>
          <div className="flex items-center justify-center p-4 bg-muted/30 rounded">
            <ThemeToggleDropdown />
          </div>
        </div>

        {/* Variant 3 */}
        <div className="border rounded-lg p-6 space-y-4">
          <div>
            <h3 className="font-semibold mb-2">3. Compact Icon</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Minimal bordered icon button
            </p>
          </div>
          <div className="flex items-center justify-center p-4 bg-muted/30 rounded">
            <ThemeToggleCompact />
          </div>
        </div>

        {/* Variant 4 */}
        <div className="border rounded-lg p-6 space-y-4">
          <div>
            <h3 className="font-semibold mb-2">4. Pill Toggle</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Three-way switch with light, dark, and system
            </p>
          </div>
          <div className="flex items-center justify-center p-4 bg-muted/30 rounded">
            <ThemeTogglePill />
          </div>
        </div>

        {/* Variant 5 */}
        <div className="border rounded-lg p-6 space-y-4">
          <div>
            <h3 className="font-semibold mb-2">5. Select Dropdown</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Traditional select input with theme options
            </p>
          </div>
          <div className="flex items-center justify-center p-4 bg-muted/30 rounded">
            <ThemeToggleSelect />
          </div>
        </div>

        {/* Variant 6 */}
        <div className="border rounded-lg p-6 space-y-4">
          <div>
            <h3 className="font-semibold mb-2">6. Radio Group</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Radio buttons for theme selection
            </p>
          </div>
          <div className="flex items-center justify-center p-4 bg-muted/30 rounded">
            <ThemeToggleRadio />
          </div>
        </div>

        {/* Variant 7 */}
        <div className="border rounded-lg p-6 space-y-4">
          <div>
            <h3 className="font-semibold mb-2">7. Text Button</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Shows next theme option with icon
            </p>
          </div>
          <div className="flex items-center justify-center p-4 bg-muted/30 rounded">
            <ThemeToggleText />
          </div>
        </div>

        {/* Variant 8 */}
        <div className="border rounded-lg p-6 space-y-4">
          <div>
            <h3 className="font-semibold mb-2">8. Animated Toggle</h3>
            <p className="text-sm text-muted-foreground mb-4">
              iOS-style sliding toggle with icons
            </p>
          </div>
          <div className="flex items-center justify-center p-4 bg-muted/30 rounded">
            <ThemeToggleAnimated />
          </div>
        </div>

        {/* Variant 9 */}
        <div className="border rounded-lg p-6 space-y-4">
          <div>
            <h3 className="font-semibold mb-2">9. With Tooltip</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Icon button with hover tooltip
            </p>
          </div>
          <div className="flex items-center justify-center p-4 bg-muted/30 rounded">
            <ThemeToggleTooltip />
          </div>
        </div>

        {/* Variant 10 */}
        <div className="border rounded-lg p-6 space-y-4">
          <div>
            <h3 className="font-semibold mb-2">10. Palette Style</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Palette icon with visual theme previews
            </p>
          </div>
          <div className="flex items-center justify-center p-4 bg-muted/30 rounded">
            <ThemeTogglePalette />
          </div>
        </div>
      </div>

      {/* Current Implementation */}
      <div className="mt-12 border-t pt-8">
        <h2 className="text-2xl font-semibold mb-4">Current Implementation</h2>
        <p className="text-muted-foreground mb-4">
          The current theme toggle used in the sidebar is the Classic Toggle (Variant 1).
          You can replace it with any of the variants above by importing the desired component.
        </p>
        <div className="bg-muted/30 p-4 rounded-lg">
          <code className="text-sm">
            {`// In /app/[teamSlug]/team-layout.tsx
import { ThemeTogglePill } from "@/components/theme-toggle-variants"

// Replace <ThemeToggle /> with:
<ThemeTogglePill />`}
          </code>
        </div>
      </div>
    </div>
  )
}