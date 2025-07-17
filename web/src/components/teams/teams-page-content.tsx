'use client'

import { useState, useMemo, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { Input } from '@/components/ui/input'
import { Plus, Search, ArrowUpDown, X } from 'lucide-react'
import { UserProfileDropdown } from '@/components/teams/user-profile-dropdown'
import { InvitationActions } from '@/components/teams/invitation-actions'
import { TeamsTable } from '@/components/teams/teams-table'
import { formatDistanceToNow } from 'date-fns'
import type { Team, TeamInvitation, TeamRole } from '@/lib/teams/types'

interface TeamsPageContentProps {
  teams: (Team & { userRole: TeamRole })[]
  pendingInvitations: TeamInvitation[]
}

export function TeamsPageContent({ teams, pendingInvitations }: TeamsPageContentProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [isSearching, setIsSearching] = useState(false)

  // Filter teams based on search query
  const filteredTeams = useMemo(() => {
    if (!searchQuery.trim()) return teams

    const query = searchQuery.toLowerCase()
    return teams.filter(team => 
      team.name.toLowerCase().includes(query) ||
      team.description?.toLowerCase().includes(query)
    )
  }, [teams, searchQuery])

  const handleSearchToggle = () => {
    setIsSearching(!isSearching)
    if (isSearching) {
      setSearchQuery('')
    }
  }

  const handleClearSearch = () => {
    setSearchQuery('')
  }

  // Add keyboard shortcut for search
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Check if user is not typing in an input or textarea
      if (e.key === '/' && !isSearching && 
          e.target instanceof Element && 
          !['INPUT', 'TEXTAREA'].includes(e.target.tagName)) {
        e.preventDefault()
        setIsSearching(true)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isSearching])

  return (
    <div className="min-h-screen bg-muted/30 flex flex-col">
      {/* Professional Header */}
      <div className="bg-background border-b">
        <div className="container mx-auto px-6 py-4 max-w-7xl">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <h1 className="text-xl font-semibold">Team Management</h1>
              <span className="hidden sm:inline-block text-sm text-muted-foreground px-2 py-0.5 bg-muted rounded-full">
                {teams.length} teams
              </span>
            </div>
            <div className="flex items-center gap-3">
              {isSearching ? (
                <div className="flex items-center gap-2">
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                    <Input
                      type="search"
                      placeholder="Search teams..."
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Escape') {
                          handleSearchToggle()
                        }
                      }}
                      className="pl-8 pr-8 h-8 w-[250px] text-sm"
                      autoFocus
                    />
                    {searchQuery && (
                      <button
                        onClick={handleClearSearch}
                        className="absolute right-2 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground p-0.5"
                      >
                        <X className="h-3.5 w-3.5" />
                      </button>
                    )}
                  </div>
                  <Button variant="ghost" size="sm" onClick={handleSearchToggle}>
                    Cancel
                  </Button>
                </div>
              ) : (
                <Button variant="ghost" size="sm" onClick={handleSearchToggle} className="group">
                  <Search className="h-4 w-4 mr-2" />
                  Search Teams
                  <kbd className="ml-2 pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-70 group-hover:opacity-100">
                    /
                  </kbd>
                </Button>
              )}
              <UserProfileDropdown />
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1">
        <div className="container mx-auto px-6 py-6 max-w-7xl space-y-6">
          {/* Pending Invitations Section - Hidden when searching */}
          {!searchQuery && pendingInvitations.length > 0 && (
            <div>
              <h2 className="font-semibold mb-4">Pending Invitations</h2>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {pendingInvitations.map((invitation) => (
                  <div key={invitation.id} className="bg-background border rounded-sm p-4 flex flex-col h-full">
                    <div className="flex items-start justify-between mb-3 flex-1">
                      <div className="flex-1 min-h-[4rem]">
                        <h4 className="font-medium">{invitation.team.name}</h4>
                        <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                          {invitation.team.description}
                        </p>
                      </div>
                      <span className="text-xs px-2 py-1 bg-muted rounded-sm font-medium ml-3">
                        {invitation.role}
                      </span>
                    </div>
                    <div className="text-xs text-muted-foreground mb-4">
                      Invited by {invitation.inviter.name} • {formatDistanceToNow(invitation.createdAt, { addSuffix: true })}
                    </div>
                    <InvitationActions invitationId={invitation.id} />
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* All Teams Section */}
          <div className="bg-background border rounded-sm">
            {/* Table Header */}
            <div className="border-b px-6 py-4">
              <div className="flex items-center justify-between">
                <h2 className="font-semibold">
                  {searchQuery ? `Search Results (${filteredTeams.length})` : 'All Teams'}
                </h2>
                <div className="flex items-center gap-2">
                  <Button variant="ghost" size="sm">
                    <ArrowUpDown className="h-4 w-4 mr-2" />
                    Sort
                  </Button>
                  <Button asChild size="sm">
                    <Link href="/teams/new">
                      <Plus className="h-4 w-4 mr-2" />
                      New Team
                    </Link>
                  </Button>
                </div>
              </div>
            </div>

            {/* Teams Table with Load More */}
            <TeamsTable teams={filteredTeams} initialDisplayCount={5} />
          </div>
        </div>
      </div>

      {/* Footer */}
      <footer className="border-t border-border mt-auto">
        <div className="max-w-6xl mx-auto px-6 py-12">
          <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-8">
            <div className="flex items-center gap-4">
              <p className="text-sm text-muted-foreground">
                © {new Date().getFullYear()} Kloudlite. All rights reserved.
              </p>
            </div>
            <nav className="flex flex-wrap gap-6">
              <Link href="/terms" className="text-sm text-muted-foreground hover:text-primary">
                Terms
              </Link>
              <Link href="/privacy" className="text-sm text-muted-foreground hover:text-primary">
                Privacy
              </Link>
              <Link href="/docs" className="text-sm text-muted-foreground hover:text-primary">
                Documentation
              </Link>
              <Link href="https://github.com/kloudlite/kloudlite" className="text-sm text-muted-foreground hover:text-primary">
                GitHub
              </Link>
            </nav>
          </div>
        </div>
      </footer>
    </div>
  )
}