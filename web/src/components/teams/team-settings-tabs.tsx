'use client'

import { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Settings, Users, Server } from 'lucide-react'
import { TeamGeneralSettings } from './settings/team-general-settings'
import { TeamUserManagement } from './settings/team-user-management'
import { TeamInfrastructureSettings } from './settings/team-infrastructure-settings'
import type { Team, TeamRole } from '@/lib/teams/types'

interface TeamSettingsTabsProps {
  team: Team & { userRole: TeamRole }
  isOwner: boolean
}

export function TeamSettingsTabs({ team, isOwner }: TeamSettingsTabsProps) {
  const [activeTab, setActiveTab] = useState('general')

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold">Team Settings</h1>
        <p className="text-muted-foreground mt-1">
          Manage your team settings, members, and infrastructure
        </p>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
        <TabsList className="grid grid-cols-3 w-full max-w-md">
          <TabsTrigger value="general" className="flex items-center gap-2">
            <Settings className="h-4 w-4" />
            General
          </TabsTrigger>
          <TabsTrigger value="users" className="flex items-center gap-2">
            <Users className="h-4 w-4" />
            Users
          </TabsTrigger>
          <TabsTrigger value="infrastructure" className="flex items-center gap-2">
            <Server className="h-4 w-4" />
            Infrastructure
          </TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="space-y-6">
          <TeamGeneralSettings team={team} isOwner={isOwner} />
        </TabsContent>

        <TabsContent value="users" className="space-y-6">
          <TeamUserManagement team={team} isOwner={isOwner} />
        </TabsContent>

        <TabsContent value="infrastructure" className="space-y-6">
          <TeamInfrastructureSettings team={team} isOwner={isOwner} />
        </TabsContent>
      </Tabs>
    </div>
  )
}