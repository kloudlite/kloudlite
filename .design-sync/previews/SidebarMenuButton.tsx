import * as React from 'react'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
} from '@kloudlite/ui'

export const ActiveAndDefault = () => (
  <SidebarProvider
    className="min-h-0 w-96 overflow-hidden rounded-md border"
    style={{ height: 440 }}
  >
    <Sidebar collapsible="none" className="h-full border-r">
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
                <SidebarMenuButton>Installations</SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  </SidebarProvider>
)

export const Sizes = () => (
  <SidebarProvider
    className="min-h-0 w-96 overflow-hidden rounded-md border"
    style={{ height: 440 }}
  >
    <Sidebar collapsible="none" className="h-full border-r">
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
          <SidebarGroupLabel>Sizes</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton size="sm">Small row</SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton isActive>Default row</SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton size="lg">Large row</SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  </SidebarProvider>
)

export const OutlineVariant = () => (
  <SidebarProvider
    className="min-h-0 w-96 overflow-hidden rounded-md border"
    style={{ height: 440 }}
  >
    <Sidebar collapsible="none" className="h-full border-r">
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
          <SidebarGroupLabel>Quick actions</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton variant="outline">New environment</SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton variant="outline">Attach cluster</SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton variant="outline">Invite teammate</SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  </SidebarProvider>
)
