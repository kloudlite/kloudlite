"use client"

import { useState } from "react"

import { Check, ChevronsUpDown, Plus, Users } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { cn } from "@/lib/utils"

interface Team {
  accountid: string
  name: string
  status: string
  role?: string
}

interface TeamSelectorProps {
  teams: Team[]
  selectedTeam: string | null
  onTeamSelect: (teamId: string | null) => void
  onCreateTeam?: () => void
}

export function TeamSelector({ teams, selectedTeam, onTeamSelect, onCreateTeam = () => window.location.href = '/teams/new' }: TeamSelectorProps) {
  const [open, setOpen] = useState(false)

  const selectedTeamData = selectedTeam ? teams.find(t => t.accountid === selectedTeam) : null

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          aria-haspopup="listbox"
          className="w-[200px] justify-between font-normal"
          onKeyDown={(e) => {
            if (e.key === "ArrowDown" || e.key === "ArrowUp") {
              e.preventDefault()
              setOpen(true)
            }
          }}
        >
          <div className="flex items-center gap-2 truncate">
            <Users className="h-4 w-4 text-muted-foreground" />
            <span className="truncate">
              {selectedTeam === null ? (
                "All teams"
              ) : selectedTeamData ? (
                selectedTeamData.name
              ) : (
                "Select team"
              )}
            </span>
          </div>
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0" align="start">
        <Command>
          <CommandInput placeholder="Search teams..." className="h-9" />
          <CommandList>
            <CommandEmpty>No teams found.</CommandEmpty>
            <CommandGroup key="all-teams-group">
              <CommandItem
                value="all-teams"
                onSelect={() => {
                  onTeamSelect(null)
                  setOpen(false)
                }}
                className="cursor-pointer"
              >
                <Users className="mr-2 h-4 w-4" />
                <span>All teams</span>
                <Check
                  className={cn(
                    "ml-auto h-4 w-4",
                    selectedTeam === null ? "opacity-100" : "opacity-0"
                  )}
                />
              </CommandItem>
            </CommandGroup>
            <CommandSeparator key="separator-1" />
            <CommandGroup key="your-teams-group" heading="Your teams">
              {teams.map((team) => (
                <CommandItem
                  key={team.accountid}
                  value={team.name}
                  onSelect={() => {
                    onTeamSelect(team.accountid)
                    setOpen(false)
                  }}
                  disabled={team.status !== "active"}
                  className="cursor-pointer"
                >
                  <div className="flex flex-1 items-center justify-between">
                    <div className="flex flex-col">
                      <span className="truncate">{team.name}</span>
                      {team.role && (
                        <span className="text-xs text-muted-foreground">{team.role}</span>
                      )}
                    </div>
                    {team.status !== "active" && (
                      <span className="ml-2 text-xs text-muted-foreground">{team.status}</span>
                    )}
                  </div>
                  <Check
                    className={cn(
                      "ml-2 h-4 w-4 shrink-0",
                      selectedTeam === team.accountid ? "opacity-100" : "opacity-0"
                    )}
                  />
                </CommandItem>
              ))}
            </CommandGroup>
            <CommandSeparator key="separator-2" />
            <CommandGroup key="create-team-group">
              <CommandItem 
                value="create-new-team"
                onSelect={() => {
                  setOpen(false)
                  onCreateTeam()
                }}
                className="cursor-pointer"
              >
                <Plus className="mr-2 h-4 w-4" />
                <span>Create new team</span>
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}