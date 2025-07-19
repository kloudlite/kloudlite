# Team Settings Components

A reusable and professional team settings system with header-based tab navigation that can be used across multiple contexts.

## Components Overview

### `TeamSettingsHeader`
The main header component with tab navigation that can be reused anywhere.

### `TeamSettingsLayout`
A complete layout wrapper that includes the header and content area.

### Individual Settings Components
- `TeamGeneralSettings` - General team configuration
- `TeamUserManagement` - User and permission management  
- `TeamInfrastructureSettings` - Infrastructure management

## Usage Examples

### 1. Full Layout (Current Implementation)
```tsx
import { TeamSettingsLayout } from '@/components/teams/settings'

export function SettingsPage({ teamname, children }) {
  return (
    <TeamSettingsLayout
      teamname={teamname}
      teamDisplayName="My Team"
      showBackButton={true}
      backHref={`/${teamname}`}
    >
      {children}
    </TeamSettingsLayout>
  )
}
```

### 2. Header Only (For Custom Layouts)
```tsx
import { TeamSettingsHeader } from '@/components/teams/settings'

export function CustomPage({ teamname }) {
  return (
    <div>
      <TeamSettingsHeader 
        teamname={teamname}
        teamDisplayName="My Team"
        showBackButton={false}
      />
      <div className="custom-content">
        {/* Your custom content */}
      </div>
    </div>
  )
}
```

### 3. In Modal/Dialog
```tsx
import { TeamSettingsHeader } from '@/components/teams/settings'

export function SettingsModal({ teamname }) {
  return (
    <Dialog>
      <DialogContent className="max-w-4xl">
        <TeamSettingsHeader 
          teamname={teamname}
          teamDisplayName="Quick Settings"
          showBackButton={false}
        />
        <div className="p-6">
          {/* Settings content */}
        </div>
      </DialogContent>
    </Dialog>
  )
}
```

### 4. In Sidebar/Drawer
```tsx
import { TeamSettingsHeader } from '@/components/teams/settings'

export function SettingsDrawer({ teamname }) {
  return (
    <Sheet>
      <SheetContent side="right" className="w-[800px]">
        <TeamSettingsHeader 
          teamname={teamname}
          showBackButton={false}
        />
        <div className="p-4">
          {/* Compact settings */}
        </div>
      </SheetContent>
    </Sheet>
  )
}
```

### 5. Different URL Structure
```tsx
// For project-level settings: /project/[id]/settings/[tab]
<TeamSettingsHeader 
  teamname={`project/${projectId}`}
  teamDisplayName="Project Settings"
  backHref={`/project/${projectId}`}
/>

// For organization settings: /org/[slug]/settings/[tab]  
<TeamSettingsHeader 
  teamname={`org/${orgSlug}`}
  teamDisplayName="Organization Settings"
  backHref={`/org/${orgSlug}`}
/>
```

## Features

### Professional Design
- ✅ Consistent with existing design system
- ✅ Hover tooltips with descriptions
- ✅ Active state indicators
- ✅ Smooth transitions and animations
- ✅ Professional spacing and typography

### Reusability
- ✅ URL-based navigation
- ✅ Configurable back button
- ✅ Flexible teamname routing
- ✅ Custom display names
- ✅ Multiple context support

### Accessibility
- ✅ Keyboard navigation
- ✅ Screen reader support
- ✅ Focus management
- ✅ ARIA attributes

## Navigation Structure

The tabs automatically generate URLs based on the `teamname` prop:

```
/{teamname}/settings/general
/{teamname}/settings/users  
/{teamname}/settings/infrastructure
```

This allows for:
- Bookmarkable URLs
- Browser back/forward navigation
- Deep linking to specific settings
- Multiple settings contexts (teams, projects, orgs)

## Customization

### Colors and Styling
All components use design tokens and can be customized via CSS variables or className overrides.

### Tab Configuration
The tabs are configured in `team-settings-header.tsx` and can be modified to add/remove sections:

```tsx
const settingsNavigation = [
  {
    name: 'General',
    href: '/settings/general',
    icon: Settings,
    description: 'Basic information and settings'
  },
  // Add more tabs here
]
```

### Layout Customization
The layout component accepts a `className` prop for custom styling:

```tsx
<TeamSettingsLayout 
  className="max-w-4xl" // Custom max-width
  // other props...
>
```