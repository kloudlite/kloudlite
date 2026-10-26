import * as React from 'react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
} from '@kloudlite/ui'

export const ResizeRail = () => (
  <SidebarProvider
    className="min-h-0 w-96 overflow-hidden rounded-md border"
    style={{ height: 440 }}
  >
<div className="relative flex h-full">
    <Sidebar collapsible="none" className="h-full w-64">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg">
              <div className="flex h-6 w-6 items-center justify-center rounded-md bg-primary text-xs font-medium text-primary-foreground">
                K
              </div>
              <div className="flex flex-col">
                <span className="text-sm font-medium">Kloudlite</span>
                <span className="text-xs text-muted-foreground">acme-platform</span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Platform</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton isActive>Environments</SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton>Clusters</SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton>Logs</SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton>
              <span className="text-sm">karthik@kloudlite.io</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
      <SidebarRail className="right-0 bg-sidebar-border" />
    </div>
    <SidebarInset className="p-4">
      <p className="text-sm text-muted-foreground">
        Drag the rail to toggle the navigation panel.
      </p>
    </SidebarInset>
  </SidebarProvider>
)

export const RailWithInset = () => (
  <SidebarProvider
    className="min-h-0 w-96 overflow-hidden rounded-md border"
    style={{ height: 440 }}
  >
<div className="relative flex h-full">
    <Sidebar collapsible="none" className="h-full w-64">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg">
              <div className="flex h-6 w-6 items-center justify-center rounded-md bg-primary text-xs font-medium text-primary-foreground">
                K
              </div>
              <div className="flex flex-col">
                <span className="text-sm font-medium">Kloudlite</span>
                <span className="text-xs text-muted-foreground">acme-platform</span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Workspace</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton isActive>Logs</SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton>Team</SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton>Settings</SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
      <SidebarRail className="right-0 bg-sidebar-border" />
    </div>
    <SidebarInset className="p-4">
      <h3 className="text-sm font-medium">Logs</h3>
    </SidebarInset>
  </SidebarProvider>
)
