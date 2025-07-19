import { TeamSettingsHeader } from './team-settings-header'

interface StandaloneSettingsProps {
  teamname: string
  teamDisplayName: string
  children?: React.ReactNode
}

// Example component showing how to use the settings header in other contexts
export function StandaloneSettings({ 
  teamname, 
  teamDisplayName, 
  children 
}: StandaloneSettingsProps) {
  return (
    <div className="min-h-screen">
      {/* Reusable header component with breadcrumbs and tabs - Full width */}
      <TeamSettingsHeader 
        teamname={teamname}
        teamDisplayName={teamDisplayName}
      />
      
      {/* Custom content area - Containerized */}
      <div className="container mx-auto px-6 py-8 max-w-7xl">
        {children || (
          <div className="text-center py-12">
            <h2 className="text-lg font-semibold mb-2">Settings Content</h2>
            <p className="text-muted-foreground">
              This shows how the settings header can be reused in different contexts.
            </p>
          </div>
        )}
      </div>
    </div>
  )
}

// Example usage in other components:
/*
import { StandaloneSettings } from '@/components/teams/settings/standalone-settings-example'

export function ProjectSettings() {
  return (
    <StandaloneSettings 
      teamname="my-team"
      teamDisplayName="My Team"
    >
      <div>Project-specific settings content here...</div>
    </StandaloneSettings>
  )
}

// Or use just the header directly:
import { TeamSettingsHeader } from '@/components/teams/settings/team-settings-header'

export function AnotherPage() {
  return (
    <div>
      <TeamSettingsHeader 
        teamname="my-team"
        teamDisplayName="My Team"
        showBackButton={false}
      />
      <div className="p-6">
        Custom page content...
      </div>
    </div>
  )
}
*/