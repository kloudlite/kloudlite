"use client"

import * as React from "react"
import { Moon, Sun, Monitor, Palette, CircleHalf, Laptop } from "lucide-react"
import { useTheme } from "next-themes"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { cn } from "@/lib/utils"

// Variant 1: Classic Toggle Button (Sun/Moon)
export function ThemeToggleClassic() {
  const { theme, setTheme } = useTheme()

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={() => setTheme(theme === "light" ? "dark" : "light")}
      className="h-9 w-9"
    >
      <Sun className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
      <Moon className="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
      <span className="sr-only">Toggle theme</span>
    </Button>
  )
}

// Variant 2: Dropdown with Icons
export function ThemeToggleDropdown() {
  const { setTheme } = useTheme()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="h-9 w-9">
          <Sun className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
          <Moon className="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
          <span className="sr-only">Toggle theme</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => setTheme("light")}>
          <Sun className="mr-2 h-4 w-4" />
          Light
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setTheme("dark")}>
          <Moon className="mr-2 h-4 w-4" />
          Dark
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setTheme("system")}>
          <Monitor className="mr-2 h-4 w-4" />
          System
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

// Variant 3: Compact Icon-only Switcher
export function ThemeToggleCompact() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return null

  return (
    <button
      onClick={() => setTheme(theme === "light" ? "dark" : "light")}
      className="relative h-7 w-7 rounded-md border bg-background p-1.5 transition-colors hover:bg-muted"
      aria-label="Toggle theme"
    >
      {theme === "light" ? (
        <Moon className="h-full w-full" />
      ) : (
        <Sun className="h-full w-full" />
      )}
    </button>
  )
}

// Variant 4: Pill Toggle Switch
export function ThemeTogglePill() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return null

  return (
    <div className="flex items-center gap-1 rounded-full border bg-muted p-0.5">
      <button
        onClick={() => setTheme("light")}
        className={cn(
          "rounded-full p-1.5 transition-colors",
          theme === "light" ? "bg-background shadow-sm" : "hover:bg-background/50"
        )}
        aria-label="Light mode"
      >
        <Sun className="h-3.5 w-3.5" />
      </button>
      <button
        onClick={() => setTheme("dark")}
        className={cn(
          "rounded-full p-1.5 transition-colors",
          theme === "dark" ? "bg-background shadow-sm" : "hover:bg-background/50"
        )}
        aria-label="Dark mode"
      >
        <Moon className="h-3.5 w-3.5" />
      </button>
      <button
        onClick={() => setTheme("system")}
        className={cn(
          "rounded-full p-1.5 transition-colors",
          theme === "system" ? "bg-background shadow-sm" : "hover:bg-background/50"
        )}
        aria-label="System theme"
      >
        <Monitor className="h-3.5 w-3.5" />
      </button>
    </div>
  )
}

// Variant 5: Select Dropdown
export function ThemeToggleSelect() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return null

  return (
    <Select value={theme} onValueChange={setTheme}>
      <SelectTrigger className="h-9 w-[110px]">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="light">
          <div className="flex items-center gap-2">
            <Sun className="h-3.5 w-3.5" />
            Light
          </div>
        </SelectItem>
        <SelectItem value="dark">
          <div className="flex items-center gap-2">
            <Moon className="h-3.5 w-3.5" />
            Dark
          </div>
        </SelectItem>
        <SelectItem value="system">
          <div className="flex items-center gap-2">
            <Monitor className="h-3.5 w-3.5" />
            System
          </div>
        </SelectItem>
      </SelectContent>
    </Select>
  )
}

// Variant 6: Radio Group
export function ThemeToggleRadio() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return null

  return (
    <RadioGroup value={theme} onValueChange={setTheme} className="flex gap-4">
      <div className="flex items-center gap-2">
        <RadioGroupItem value="light" id="light" />
        <Label htmlFor="light" className="flex items-center gap-1 cursor-pointer">
          <Sun className="h-3.5 w-3.5" />
          Light
        </Label>
      </div>
      <div className="flex items-center gap-2">
        <RadioGroupItem value="dark" id="dark" />
        <Label htmlFor="dark" className="flex items-center gap-1 cursor-pointer">
          <Moon className="h-3.5 w-3.5" />
          Dark
        </Label>
      </div>
      <div className="flex items-center gap-2">
        <RadioGroupItem value="system" id="system" />
        <Label htmlFor="system" className="flex items-center gap-1 cursor-pointer">
          <Monitor className="h-3.5 w-3.5" />
          System
        </Label>
      </div>
    </RadioGroup>
  )
}

// Variant 7: Minimal Text Button
export function ThemeToggleText() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return null

  const nextTheme = theme === "light" ? "dark" : theme === "dark" ? "system" : "light"
  const nextLabel = nextTheme.charAt(0).toUpperCase() + nextTheme.slice(1)

  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={() => setTheme(nextTheme)}
      className="h-8 px-2 text-xs"
    >
      {theme === "light" && <Sun className="mr-1.5 h-3.5 w-3.5" />}
      {theme === "dark" && <Moon className="mr-1.5 h-3.5 w-3.5" />}
      {theme === "system" && <Monitor className="mr-1.5 h-3.5 w-3.5" />}
      {nextLabel}
    </Button>
  )
}

// Variant 8: Animated Toggle
export function ThemeToggleAnimated() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return null

  return (
    <button
      onClick={() => setTheme(theme === "light" ? "dark" : "light")}
      className="group relative h-9 w-16 rounded-full bg-muted p-1 transition-colors hover:bg-muted/80"
      aria-label="Toggle theme"
    >
      <div
        className={cn(
          "absolute h-7 w-7 rounded-full bg-background shadow-sm transition-all duration-300",
          theme === "dark" ? "translate-x-7" : "translate-x-0"
        )}
      >
        <div className="flex h-full w-full items-center justify-center">
          {theme === "light" ? (
            <Sun className="h-4 w-4 text-yellow-500" />
          ) : (
            <Moon className="h-4 w-4 text-blue-500" />
          )}
        </div>
      </div>
    </button>
  )
}

// Variant 9: Icon with Tooltip (requires tooltip component)
export function ThemeToggleTooltip() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return null

  const nextTheme = theme === "light" ? "dark" : theme === "dark" ? "system" : "light"

  return (
    <div className="group relative">
      <Button
        variant="ghost"
        size="icon"
        onClick={() => setTheme(nextTheme)}
        className="h-9 w-9"
      >
        {theme === "light" && <Sun className="h-5 w-5" />}
        {theme === "dark" && <Moon className="h-5 w-5" />}
        {theme === "system" && <Monitor className="h-5 w-5" />}
      </Button>
      <div className="pointer-events-none absolute -top-8 left-1/2 -translate-x-1/2 opacity-0 transition-opacity group-hover:opacity-100">
        <div className="rounded bg-popover px-2 py-1 text-xs text-popover-foreground shadow-md">
          Switch to {nextTheme}
        </div>
      </div>
    </div>
  )
}

// Variant 10: Palette Style
export function ThemeTogglePalette() {
  const { setTheme } = useTheme()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="h-9 w-9">
          <Palette className="h-5 w-5" />
          <span className="sr-only">Change theme</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-40">
        <DropdownMenuItem onClick={() => setTheme("light")} className="gap-3">
          <div className="flex h-5 w-5 items-center justify-center rounded-full bg-white border">
            <Sun className="h-3 w-3" />
          </div>
          Light
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setTheme("dark")} className="gap-3">
          <div className="flex h-5 w-5 items-center justify-center rounded-full bg-slate-900 border">
            <Moon className="h-3 w-3 text-white" />
          </div>
          Dark
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setTheme("system")} className="gap-3">
          <div className="flex h-5 w-5 items-center justify-center rounded-full bg-gradient-to-r from-white to-slate-900 border">
            <CircleHalf className="h-3 w-3" />
          </div>
          System
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}