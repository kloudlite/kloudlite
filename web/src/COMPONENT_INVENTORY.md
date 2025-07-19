# Component Inventory

## UI Components (`/src/components/ui/`)

### Core UI Components
- **alert-dialog.tsx** - Modal dialogs for confirmations and alerts
- **avatar.tsx** - User avatar display component
- **badge.tsx** - Status and category badges
- **breadcrumb.tsx** - Navigation breadcrumb component
- **button.tsx** - Primary button component with variants
- **card.tsx** - Container cards for content sections
- **checkbox.tsx** - Form checkbox input
- **data-table.tsx** - Sortable, filterable data table
- **form.tsx** - Form components with validation
- **input.tsx** - Text input field component
- **label.tsx** - Form field labels
- **link.tsx** - Navigation link component
- **radio-group.tsx** - Radio button group input
- **select.tsx** - Dropdown select component
- **separator.tsx** - Visual divider component
- **switch.tsx** - Toggle switch component
- **table.tsx** - Basic table components
- **textarea.tsx** - Multi-line text input
- **use-toast.tsx** - Toast notification system

### Newly Created Standardized Components
- **section.tsx** - Standardized section containers
- **overview-card.tsx** - Metric display cards and grids

## Layout Components (`/src/components/layout/`)

### Layout System
- **page-container.tsx** - Page-level layout containers
- **grid-layout.tsx** - Grid layout utilities

## Dashboard Components (`/src/components/dashboard/`)

### Dashboard Layout
- **layout/dashboard-layout.tsx** - Main dashboard layout wrapper

### Dashboard Features
- **AdminDashboard.tsx** - Admin dashboard view
- **activity/ActivityFeed.tsx** - Activity feed component
- **stats/CleanNodePoolGrid.tsx** - Node pool statistics grid
- **stats/CleanWorkMachineTable.tsx** - Work machine table display
- **stats/NodePoolStats.tsx** - Node pool metrics

## Team Components (`/src/components/teams/`)

### Team Management
- **team-card.tsx** - Individual team display card
- **teams-table.tsx** - Teams listing table
- **teams-page-content.tsx** - Main teams page content
- **team-members-list.tsx** - Team member listing
- **team-settings.tsx** - Team settings form
- **team-settings-tabs.tsx** - Settings navigation tabs
- **user-profile-dropdown.tsx** - User profile menu

### Team Invitations
- **invite-member-dialog.tsx** - Member invitation modal
- **team-invitation-card.tsx** - Invitation display card
- **compact-invitation-card.tsx** - Compact invitation view
- **invitation-actions.tsx** - Invitation action buttons

### Team Forms
- **create-team-form.tsx** - New team creation form

### Team Settings (`/src/components/teams/settings/`)
- **team-settings-layout.tsx** - Settings page layout
- **team-settings-header.tsx** - Settings header with navigation
- **team-general-settings.tsx** - General team settings
- **team-user-management.tsx** - User management interface
- **team-infrastructure-settings.tsx** - Infrastructure configuration
- **standalone-settings-example.tsx** - Example settings component

## Specialized Components

### Action Grid (`/src/components/ui/action-grid.tsx`)
- Quick action buttons and metric cards

## Component Patterns Identified

### ✅ Standardized Patterns
1. **Section Headers**: Now use `CompleteSection` component
2. **Overview Cards**: Now use `OverviewCard` and `OverviewGrid`
3. **Layout Containers**: Use `LAYOUT.*` constants

### 🔄 Patterns Needing Standardization
1. **Form Layouts**: Repeated form grid patterns across settings
2. **Table Headers**: Similar header patterns in multiple tables
3. **Card Grids**: Various grid layouts that could use standard patterns
4. **Mobile/Desktop Responsive Tables**: Repeated mobile card + desktop table pattern

### 🚨 Redundant Components
1. **team-settings.tsx** vs **team-settings/*** - May have overlapping functionality
2. **team-invitation-card.tsx** vs **compact-invitation-card.tsx** - Similar purpose, different sizes
3. **teams-table.tsx** vs **teams-page-content.tsx** - Table logic may be duplicated

## Consolidation Opportunities

### Form Components
- **Pattern**: Multiple settings components use similar form layouts
- **Solution**: Create standardized `SettingsForm` component with sections

### Table Components  
- **Pattern**: Mobile card + desktop table pattern repeated across components
- **Solution**: Create standardized `ResponsiveTable` component

### Card Grids
- **Pattern**: Various metric/overview grids with similar structure
- **Solution**: Expand `OverviewGrid` to support more use cases

## State Management Analysis

### Client-Side State (useState/useEffect)
- **teams-page-content.tsx**: Search state, keyboard navigation
- **team-infrastructure-settings.tsx**: Toggle states for policies
- **team-general-settings.tsx**: Form state, submission state
- **team-user-management.tsx**: Form state, member list state

### Potential Server Component Migrations
1. **Static Data**: Team settings that don't change frequently
2. **Initial States**: Default values that can be pre-rendered
3. **Configuration**: Infrastructure policies that are set server-side

## Recommendations

### Immediate Actions
1. **Complete token replacement** in remaining components
2. **Consolidate invitation card components** into single flexible component
3. **Create ResponsiveTable component** for mobile/desktop patterns
4. **Standardize form layouts** across settings components

### Medium-term Goals
1. **Move static state to server components** where possible
2. **Create component style guide** with usage examples
3. **Implement automated linting** for hardcoded values
4. **Add component testing** for standardized components

### Long-term Vision
1. **Full design system implementation** with Storybook
2. **Automated component generation** for common patterns
3. **Performance optimization** through better component reuse
4. **Accessibility standardization** across all components