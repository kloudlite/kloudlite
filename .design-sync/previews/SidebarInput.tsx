import * as React from 'react'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInput,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
} from '@kloudlite/ui'

export const SearchInHeader = () => (
  <SidebarProvider
    className="min-h-0 w-96 overflow-hidden rounded-md border"
    style={{ height: 440 }}
  >
    <Sidebar collapsible="none" className="h-full border-r">
      <SidebarHeader>
        <span className="px-2 text-sm font-medium">Kloudlite</span>
        <SidebarInput placeholder="Search environments" />
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
                <SidebarMenuButton>Deployments</SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  </SidebarProvider>
)

export const FilterWithValue = () => (
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
          <SidebarGroupLabel>Clusters</SidebarGroupLabel>
          <SidebarGroupContent className="flex flex-col gap-2">
            <SidebarInput defaultValue="eu-" />
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton isActive>eu-west-1</SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton>eu-central-1</SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  </SidebarProvider>
)
