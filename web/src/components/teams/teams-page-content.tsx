'use client'

import { useState, useEffect, useRef } from 'react'
import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { Input } from '@/components/ui/input'
import { Plus, Search, X, Crown, Shield, UserCheck, Settings, MoreVertical } from 'lucide-react'
import { UserProfileDropdown } from '@/components/teams/user-profile-dropdown'
import { InvitationActions } from '@/components/teams/invitation-actions'
import { formatDistanceToNow } from 'date-fns'
import type { Team, TeamInvitation, TeamRole } from '@/lib/teams/types'
import { PageContainer, PageHeader, PageContent, PageSection } from '@/components/layout/page-container'
import { GridLayout, Card } from '@/components/layout/grid-layout'
import { DataTable, type ColumnDef } from '@/components/ui/data-table'
import { useTableState } from '@/hooks/use-table-state'

interface TeamsPageContentProps {
  teams: (Team & { userRole: TeamRole })[]
  pendingInvitations: TeamInvitation[]
}

const roleIcons = {
  owner: Crown,
  admin: Shield,
  member: UserCheck,
}

export function TeamsPageContent({ teams, pendingInvitations }: TeamsPageContentProps) {
  const [isSearching, setIsSearching] = useState(false)
  const [focusedRowIndex, setFocusedRowIndex] = useState(-1)
  const searchInputRef = useRef<HTMLInputElement>(null)
  const tableRef = useRef<HTMLDivElement>(null)

  // Use our reusable table state hook
  const {
    processedData: filteredTeams,
    searchQuery,
    handleSearch,
    clearSearch,
    sortField,
    sortDirection,
    handleSort,
  } = useTableState({
    data: teams,
    searchFields: ['name', 'description', 'userRole', 'region'],
    initialSort: { field: 'name', direction: 'asc' }
  })

  const handleSearchToggle = () => {
    setIsSearching(!isSearching)
    if (isSearching) {
      clearSearch()
      setFocusedRowIndex(-1)
    }
  }

  const handleKeyNavigation = (e: React.KeyboardEvent) => {
    if (!isSearching) return

    if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (focusedRowIndex === -1) {
        // Move from search input to first table row
        setFocusedRowIndex(0)
        searchInputRef.current?.blur()
      } else if (focusedRowIndex < filteredTeams.length - 1) {
        setFocusedRowIndex(prev => prev + 1)
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (focusedRowIndex > 0) {
        setFocusedRowIndex(prev => prev - 1)
      } else if (focusedRowIndex === 0) {
        // Move back to search input
        setFocusedRowIndex(-1)
        searchInputRef.current?.focus()
      }
    } else if (e.key === 'Enter' && focusedRowIndex >= 0) {
      e.preventDefault()
      // Navigate to the focused team
      const team = filteredTeams[focusedRowIndex]
      if (team) {
        window.location.href = `/${team.slug || team.name.toLowerCase().replace(/\s+/g, '-')}`
      }
    }
  }

  // Add keyboard shortcut for search and global navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Check if user is not typing in an input or textarea
      if (e.key === '/' && !isSearching && 
          e.target instanceof Element && 
          !['INPUT', 'TEXTAREA'].includes(e.target.tagName)) {
        e.preventDefault()
        setIsSearching(true)
        return
      }

      // Handle global navigation when searching
      if (isSearching) {
        if (e.key === 'ArrowDown') {
          e.preventDefault()
          if (focusedRowIndex === -1) {
            setFocusedRowIndex(0)
            searchInputRef.current?.blur()
          } else if (focusedRowIndex < filteredTeams.length - 1) {
            setFocusedRowIndex(prev => prev + 1)
          }
        } else if (e.key === 'ArrowUp') {
          e.preventDefault()
          if (focusedRowIndex > 0) {
            setFocusedRowIndex(prev => prev - 1)
          } else if (focusedRowIndex === 0) {
            setFocusedRowIndex(-1)
            searchInputRef.current?.focus()
          }
        } else if (e.key === 'Enter' && focusedRowIndex >= 0) {
          e.preventDefault()
          const team = filteredTeams[focusedRowIndex]
          if (team) {
            window.location.href = `/${team.slug || team.name.toLowerCase().replace(/\s+/g, '-')}`
          }
        } else if (e.key === 'Escape') {
          e.preventDefault()
          handleSearchToggle()
        }
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isSearching, focusedRowIndex, filteredTeams, handleSearchToggle])

  // Reset focus when search results change
  useEffect(() => {
    setFocusedRowIndex(-1)
  }, [searchQuery])

  // Define table columns
  const columns: ColumnDef<Team & { userRole: TeamRole }>[] = [
    {
      key: 'name',
      header: 'Team Name',
      sortable: true,
      render: (_, team) => (
        <div>
          <Link 
            href={`/${team.slug || team.name.toLowerCase().replace(/\s+/g, '-')}`} 
            className="font-medium hover:text-primary transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 rounded-sm"
            aria-label={`Go to ${team.name} team dashboard`}
          >
            {team.name}
          </Link>
          <div className="text-sm text-muted-foreground mt-0.5">
            <p className="max-w-md truncate">{team.description}</p>
            {/* Show role and member count on mobile */}
            <div className="flex items-center gap-4 mt-1 sm:hidden">
              <div className="flex items-center gap-1.5">
                {roleIcons[team.userRole] && 
                  (() => {
                    const IconComponent = roleIcons[team.userRole];
                    return <IconComponent className="h-3.5 w-3.5" />;
                  })()
                }
                <span className="capitalize">{team.userRole}</span>
              </div>
              <span>{team.memberCount} members</span>
            </div>
          </div>
        </div>
      )
    },
    {
      key: 'userRole',
      header: 'Role',
      sortable: true,
      className: 'hidden sm:table-cell',
      render: (_, team) => (
        <div className="flex items-center gap-1.5">
          {roleIcons[team.userRole] && 
            (() => {
              const IconComponent = roleIcons[team.userRole];
              return <IconComponent className="h-3.5 w-3.5 text-muted-foreground" />;
            })()
          }
          <span className="text-sm capitalize">{team.userRole}</span>
        </div>
      )
    },
    {
      key: 'memberCount',
      header: 'Members',
      sortable: true,
      className: 'hidden md:table-cell',
      render: (value) => <span className="text-sm">{value}</span>
    },
    {
      key: 'lastActivity',
      header: 'Last Accessed',
      sortable: true,
      className: 'hidden lg:table-cell',
      render: (value) => (
        <span className="text-sm text-muted-foreground">
          {value ? formatDistanceToNow(value as Date, { addSuffix: true }) : 'Never'}
        </span>
      )
    },
    {
      key: 'joinedAt',
      header: 'Member Since',
      sortable: true,
      className: 'hidden lg:table-cell',
      render: (_, team) => (
        <span className="text-sm text-muted-foreground">
          {team.userRole === 'owner' 
            ? formatDistanceToNow(team.createdAt, { addSuffix: true })
            : formatDistanceToNow(team.joinedAt || team.createdAt, { addSuffix: true })
          }
        </span>
      )
    },
    {
      key: 'id',
      header: '',
      className: 'w-16',
      render: (_, team) => (
        <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity">
          <Button variant="ghost" size="icon-sm" asChild>
            <Link 
              href={`/${team.slug || team.name.toLowerCase().replace(/\s+/g, '-')}/settings`}
              aria-label={`${team.name} settings`}
            >
              <Settings className="h-4 w-4" />
            </Link>
          </Button>
          <Button 
            variant="ghost" 
            size="icon-sm"
            aria-label={`More actions for ${team.name}`}
          >
            <MoreVertical className="h-4 w-4" />
          </Button>
        </div>
      )
    }
  ]

  return (
    <PageContainer>
      <div className="relative">
        {/* Mobile: Hide header completely in search mode */}
        <div className={isSearching ? "hidden sm:block" : ""}>
          <PageHeader
            title="Teams"
            description="Manage your team memberships and invitations"
            actions={
              // Show actions only when not searching or on desktop
              !isSearching ? (
                <>
                  <Button variant="ghost" size="sm" onClick={handleSearchToggle} className="group">
                    <Search className="h-4 w-4 mr-2" />
                    Search Teams
                    <kbd 
                      className="ml-2 pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-70 group-hover:opacity-100"
                      aria-label="Press forward slash to search"
                    >
                      /
                    </kbd>
                  </Button>
                  <UserProfileDropdown 
                    user={{
                      name: "John Doe",
                      email: "john.doe@example.com"
                    }}
                  />
                </>
              ) : undefined
            }
          />
        </div>
        
        {/* Mobile: Search-only header when searching */}
        {isSearching && (
          <div className="sm:hidden bg-background border-b">
            <div className="px-4 py-3">
              <div className="flex items-center gap-2">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                  <Input
                    ref={searchInputRef}
                    type="text"
                    placeholder="Search teams, roles, regions..."
                    value={searchQuery}
                    onChange={(e) => handleSearch(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'Enter' || e.key === 'Escape') {
                        return
                      }
                    }}
                    className="pl-8 pr-8 h-9 text-sm"
                    autoFocus
                  />
                  {searchQuery && (
                    <button
                      onClick={clearSearch}
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
            </div>
          </div>
        )}
        
        {/* Desktop: Regular header with search actions */}
        {isSearching && (
          <div className="hidden sm:block bg-background border-b">
            <div className="max-w-6xl mx-auto px-6 py-4">
              <div className="flex items-center justify-end gap-2">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                  <Input
                    ref={searchInputRef}
                    type="text"
                    placeholder="Search teams, roles, regions..."
                    value={searchQuery}
                    onChange={(e) => handleSearch(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'Enter' || e.key === 'Escape') {
                        return
                      }
                    }}
                    className="pl-8 pr-8 h-8 w-[280px] text-sm"
                    autoFocus
                  />
                  {searchQuery && (
                    <button
                      onClick={clearSearch}
                      className="absolute right-2 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground p-0.5"
                    >
                      <X className="h-3.5 w-3.5" />
                    </button>
                  )}
                </div>
                <Button variant="ghost" size="sm" onClick={handleSearchToggle}>
                  Cancel
                </Button>
                <UserProfileDropdown 
                  user={{
                    name: "John Doe",
                    email: "john.doe@example.com"
                  }}
                />
              </div>
            </div>
          </div>
        )}
      </div>

      <PageContent>
        {/* Pending Invitations Section */}
        {!searchQuery && pendingInvitations.length > 0 && (
          <div className="mb-8">
            <div className="mb-6">
              <h2 className="text-xl font-semibold mb-2">Pending Invitations</h2>
              <p className="text-sm text-muted-foreground">Review and respond to team invitations</p>
            </div>
            <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
              {pendingInvitations.map((invitation) => (
                <div key={invitation.id} className="bg-background border rounded-lg p-4 hover:shadow-sm transition-shadow">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex-1 min-w-0">
                      <h3 className="font-semibold text-base truncate">{invitation.team.name}</h3>
                      <div className="flex items-center gap-1.5 mt-1">
                        {invitation.role === 'owner' && <Crown className="h-3 w-3 text-amber-500" />}
                        {invitation.role === 'admin' && <Shield className="h-3 w-3 text-blue-500" />}
                        {invitation.role === 'member' && <UserCheck className="h-3 w-3 text-green-500" />}
                        <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                          {invitation.role}
                        </span>
                      </div>
                    </div>
                    <InvitationActions invitationId={invitation.id} />
                  </div>
                  <div className="text-xs text-muted-foreground">
                    by <span className="font-medium">{invitation.inviter.name}</span> • {formatDistanceToNow(invitation.createdAt, { addSuffix: true })}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* All Teams Section */}
        <div className="bg-background border rounded-lg overflow-hidden">
          {/* Table Header */}
          <div className="border-b px-6 py-4">
            <div className="flex items-center justify-between">
              <h2 className="font-semibold">
                {searchQuery ? `Search Results (${filteredTeams.length})` : 'All Teams'}
              </h2>
              <div className="flex items-center gap-2">
                <Button asChild size="sm">
                  <Link href="/teams/new">
                    <Plus className="h-4 w-4 mr-2" />
                    New Team
                  </Link>
                </Button>
              </div>
            </div>
          </div>

          {/* Teams Table */}
          <DataTable
            data={filteredTeams}
            columns={columns}
            initialDisplayCount={5}
            onSort={handleSort}
            sortField={sortField}
            sortDirection={sortDirection}
            focusedRowIndex={focusedRowIndex}
            emptyState={{
              title: 'No teams found',
              description: searchQuery 
                ? 'Try adjusting your search to find teams'
                : 'Get started by creating your first team',
              action: !searchQuery ? (
                <Button size="sm" asChild>
                  <Link href="/teams/new">
                    <Plus className="h-4 w-4 mr-2" />
                    Create Team
                  </Link>
                </Button>
              ) : undefined
            }}
          />
        </div>
      </PageContent>

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
    </PageContainer>
  )
}