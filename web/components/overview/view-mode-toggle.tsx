"use client"

import { LayoutGrid, List } from "lucide-react"

import { Button } from "@/components/ui/button"

interface ViewModeToggleProps {
  viewMode: "grid" | "list"
  onViewModeChange: (mode: "grid" | "list") => void
}

export function ViewModeToggle({ viewMode, onViewModeChange }: ViewModeToggleProps) {
  return (
    <div className="flex rounded-lg border p-1">
      <Button
        variant={viewMode === "grid" ? "secondary" : "ghost"}
        size="icon"
        className="h-8 w-8"
        onClick={() => onViewModeChange("grid")}
        aria-label="Grid view"
      >
        <LayoutGrid className="h-4 w-4" />
      </Button>
      <Button
        variant={viewMode === "list" ? "secondary" : "ghost"}
        size="icon"
        className="h-8 w-8"
        onClick={() => onViewModeChange("list")}
        aria-label="List view"
      >
        <List className="h-4 w-4" />
      </Button>
    </div>
  )
}