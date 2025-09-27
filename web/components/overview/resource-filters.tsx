"use client"

import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"

interface ResourceFiltersProps {
  activeTab: string
  onTabChange: (value: string) => void
}

export function ResourceFilters({ activeTab, onTabChange }: ResourceFiltersProps) {
  return (
    <Tabs value={activeTab} onValueChange={onTabChange} className="w-full sm:w-auto">
      <TabsList className="grid w-full grid-cols-3 bg-muted/50 sm:inline-flex sm:w-auto">
        <TabsTrigger 
          value="all" 
          className="data-[state=active]:bg-background data-[state=active]:shadow-sm"
        >
          All
        </TabsTrigger>
        <TabsTrigger 
          value="environment" 
          className="data-[state=active]:bg-background data-[state=active]:shadow-sm"
        >
          Environments
        </TabsTrigger>
        <TabsTrigger 
          value="workspace" 
          className="data-[state=active]:bg-background data-[state=active]:shadow-sm"
        >
          Workspaces
        </TabsTrigger>
      </TabsList>
    </Tabs>
  )
}