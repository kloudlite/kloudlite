"use client"

import { useState } from "react"

import { Plus, Search } from "lucide-react"
import { useRouter, useSearchParams } from "next/navigation"

import { ResourceCard } from "@/components/resource-card"
import { TeamSelector } from "@/components/team-selector"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

import { CreateResourceMenu } from "./create-resource-menu"
import { ResourceFilters } from "./resource-filters"
import { ViewModeToggle } from "./view-mode-toggle"

interface Resource {
  id: string
  type: "environment" | "workspace"
  name: string
  team: string
  teamId: string
  status: "active" | "inactive" | "deploying"
  description?: string
  lastUpdated?: string
}

interface ResourceListWrapperProps {
  resources: Resource[]
  selectedTeam: string | null
  searchQuery: string
  teams?: Array<{
    teamId: string
    displayName: string
    status?: string
    role?: string
  }>
}

export function ResourceListWrapper({ resources, selectedTeam, searchQuery, teams = [] }: ResourceListWrapperProps) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [activeTab, setActiveTab] = useState(searchParams.get("tab") || "all")
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid")
  const [mobileSearchQuery, setMobileSearchQuery] = useState(searchQuery)
  
  // Map teams to the expected format for mobile
  const formattedTeams = teams.map(team => ({
    accountid: team.teamId,
    name: team.displayName,
    status: team.status || 'active',
    role: team.role
  }))

  // Filter resources based on selected team, search query, and active tab
  const filteredResources = resources.filter(resource => {
    const matchesTeam = !selectedTeam || resource.teamId === selectedTeam
    const matchesSearch = !searchQuery || 
      resource.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      resource.description?.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesTab = activeTab === "all" || resource.type === activeTab
    
    return matchesTeam && matchesSearch && matchesTab
  })

  const handleCreateResource = (_type: "environment" | "workspace") => {
    // TODO: Implement actual creation logic
  }

  const handleTabChange = (value: string) => {
    setActiveTab(value)
    const params = new URLSearchParams(searchParams)
    if (value === "all") {
      params.delete("tab")
    } else {
      params.set("tab", value)
    }
    router.push(`?${params.toString()}`)
  }

  const handleTeamSelect = (teamId: string | null) => {
    const params = new URLSearchParams(searchParams)
    if (teamId) {
      params.set('team', teamId)
    } else {
      params.delete('team')
    }
    router.push(`?${params.toString()}`)
  }

  const handleMobileSearch = (e: React.FormEvent) => {
    e.preventDefault()
    const params = new URLSearchParams(searchParams)
    if (mobileSearchQuery) {
      params.set('search', mobileSearchQuery)
    } else {
      params.delete('search')
    }
    router.push(`?${params.toString()}`)
  }

  return (
    <>
      {/* Mobile Search and Team Selector */}
      <div className="space-y-4 md:hidden mb-6">
        {/* Mobile Search */}
        <form onSubmit={handleMobileSearch}>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search resources..."
              value={mobileSearchQuery}
              onChange={(e) => setMobileSearchQuery(e.target.value)}
              className="w-full pl-9"
            />
          </div>
        </form>
        
        {/* Mobile Team Selector */}
        <TeamSelector
          teams={formattedTeams}
          selectedTeam={selectedTeam}
          onTeamSelect={handleTeamSelect}
        />
      </div>

      {/* Tabs and Actions Row */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between mb-6">
        <ResourceFilters activeTab={activeTab} onTabChange={handleTabChange} />
        
        <div className="flex items-center justify-between gap-2 sm:justify-end">
          <CreateResourceMenu onCreateResource={handleCreateResource} />
          <ViewModeToggle viewMode={viewMode} onViewModeChange={setViewMode} />
        </div>
      </div>

      {/* Resources Grid/List */}
      <div>
        {teams.length === 0 ? (
          <div className="flex min-h-[300px] flex-col items-center justify-center rounded-lg border-2 border-dashed bg-muted/20 p-6 text-center md:min-h-[400px] md:p-8">
            <div className="mx-auto max-w-md">
              <h3 className="text-base font-semibold md:text-lg">Welcome to Kloudlite!</h3>
              <p className="mt-2 text-xs text-muted-foreground md:text-sm">
                You haven't joined any teams yet. Create your first team to start deploying environments and workspaces.
              </p>
              <div className="mt-4 md:mt-6">
                <Button 
                  onClick={() => router.push('/teams/new')}
                  className="group shadow-sm hover:shadow-md transition-all duration-200"
                  size="sm"
                >
                  <Plus className="mr-2 h-3.5 w-3.5 transition-transform group-hover:rotate-90 md:h-4 md:w-4" />
                  Create Your First Team
                </Button>
              </div>
            </div>
          </div>
        ) : filteredResources.length === 0 ? (
          <div className="flex min-h-[300px] flex-col items-center justify-center rounded-lg border-2 border-dashed bg-muted/20 p-6 text-center md:min-h-[400px] md:p-8">
            <div className="mx-auto max-w-md">
              <h3 className="text-base font-semibold md:text-lg">No resources found</h3>
              <p className="mt-2 text-xs text-muted-foreground md:text-sm">
                {searchQuery 
                  ? "Try adjusting your search or filters to find what you're looking for." 
                  : "Get started by creating your first environment or workspace."}
              </p>
              {!searchQuery && (
                <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-center md:mt-6 md:gap-3">
                  <Button 
                    onClick={() => handleCreateResource("environment")}
                    className="group w-full shadow-sm hover:shadow-md transition-all duration-200 sm:w-auto"
                    size="sm"
                  >
                    <Plus className="mr-2 h-3.5 w-3.5 transition-transform group-hover:rotate-90 md:h-4 md:w-4" />
                    Create Environment
                  </Button>
                  <Button 
                    variant="outline" 
                    onClick={() => handleCreateResource("workspace")}
                    className="w-full border-muted-foreground/20 hover:border-muted-foreground/40 transition-all duration-200 sm:w-auto"
                    size="sm"
                  >
                    <Plus className="mr-2 h-3.5 w-3.5 md:h-4 md:w-4" />
                    Create Workspace
                  </Button>
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className={viewMode === "grid" 
            ? "grid gap-4 md:gap-6 lg:grid-cols-2 xl:grid-cols-3" 
            : "space-y-3"
          }>
            {filteredResources.map((resource) => (
              <ResourceCard
                key={resource.id}
                type={resource.type}
                name={resource.name}
                team={resource.team}
                status={resource.status}
                description={resource.description}
                lastUpdated={resource.lastUpdated}
                onClick={() => router.push(`/resources/${resource.id}`)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Summary Stats */}
      {filteredResources.length > 0 && (
        <div className="mt-8 border-t border-border/50 pt-4 md:mt-12 md:pt-6">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground sm:gap-4 md:gap-6 md:text-sm">
              <div className="flex items-center gap-1.5 md:gap-2">
                <div className="h-1.5 w-1.5 rounded-full bg-green-500 animate-pulse md:h-2 md:w-2" />
                <span>{filteredResources.filter(r => r.status === "active").length} Active</span>
              </div>
              <div className="flex items-center gap-1.5 md:gap-2">
                <div className="h-1.5 w-1.5 rounded-full bg-yellow-500 md:h-2 md:w-2" />
                <span>{filteredResources.filter(r => r.status === "deploying").length} Deploying</span>
              </div>
              <div className="flex items-center gap-1.5 md:gap-2">
                <div className="h-1.5 w-1.5 rounded-full bg-gray-400 md:h-2 md:w-2" />
                <span>{filteredResources.filter(r => r.status === "inactive").length} Inactive</span>
              </div>
            </div>
            <p className="text-xs text-muted-foreground md:text-sm">
              Showing {filteredResources.length} of {resources.length} resources
            </p>
          </div>
        </div>
      )}
    </>
  )
}