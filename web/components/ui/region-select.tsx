"use client"

import * as React from "react"

import { Check, ChevronsUpDown } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { cn } from "@/lib/utils"

interface Region {
  id: string
  displayName: string
  value: string
}

interface RegionSelectProps {
  regions: Region[]
  value?: string
  onValueChange?: (value: string) => void
  placeholder?: string
}

export function RegionSelect({
  regions,
  value,
  onValueChange,
  placeholder = "Select a region",
}: RegionSelectProps) {
  const [open, setOpen] = React.useState(false)
  const selectedRegion = regions.find((region) => region.value === value)

  // Group regions by continent
  const groupedRegions = React.useMemo(() => {
    const groups: Record<string, Region[]> = {}
    
    regions.forEach((region) => {
      let group = "Other"
      if (region.value.startsWith("us-")) {group = "United States"}
      else if (region.value.startsWith("eu-")) {group = "Europe"}
      else if (region.value.startsWith("ap-")) {group = "Asia Pacific"}
      else if (region.value.startsWith("ca-")) {group = "Canada"}
      else if (region.value.startsWith("sa-")) {group = "South America"}
      else if (region.value.startsWith("me-") || region.value.startsWith("il-")) {group = "Middle East"}
      else if (region.value.startsWith("af-")) {group = "Africa"}
      
      if (!groups[group]) {groups[group] = []}
      groups[group].push(region)
    })
    
    return groups
  }, [regions])

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-full justify-between font-normal"
        >
          {selectedRegion ? (
            <span className="flex items-center gap-2 text-left flex-1">
              <span className="truncate">{selectedRegion.displayName}</span>
              <span className="font-mono text-xs text-muted-foreground">({selectedRegion.value})</span>
            </span>
          ) : (
            <span className="text-muted-foreground">{placeholder}</span>
          )}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[450px] p-0" align="start">
        <Command className="border-0">
          <CommandInput placeholder="Search by region ID or name..." className="h-10" />
          <CommandList className="max-h-[300px] overflow-auto">
            <CommandEmpty className="py-6 text-center text-sm">No region found.</CommandEmpty>
            {Object.entries(groupedRegions).map(([group, regions]) => (
              <CommandGroup key={group}>
                <div className="sticky top-0 bg-background z-10 px-2 py-1.5 border-b">
                  <h4 className="text-xs font-medium text-muted-foreground">{group}</h4>
                </div>
                {regions.map((region) => (
                  <CommandItem
                    key={region.value}
                    value={`${region.value} ${region.displayName}`}
                    onSelect={() => {
                      onValueChange?.(region.value)
                      setOpen(false)
                    }}
                    className="py-2 px-2"
                  >
                    <div className="flex items-center justify-between w-full gap-3">
                      <span className="text-sm truncate flex-1">
                        {region.displayName}
                      </span>
                      <span className="font-mono text-xs text-muted-foreground">
                        {region.value}
                      </span>
                      <Check
                        className={cn(
                          "h-4 w-4 shrink-0 ml-2",
                          value === region.value ? "opacity-100" : "opacity-0"
                        )}
                      />
                    </div>
                  </CommandItem>
                ))}
              </CommandGroup>
            ))}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}