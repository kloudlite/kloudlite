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
  SidebarTrigger,
} from '@kloudlite/ui'

export const InInsetHeader = () => (
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
    <SidebarInset>
      <div className="flex items-center gap-2 border-b p-4">
        <SidebarTrigger />
        <span className="text-sm font-medium">Environments</span>
      </div>
    </SidebarInset>
  </SidebarProvider>
)

export const InSidebarHeader = () => (
  <SidebarProvider
    className="min-h-0 w-96 overflow-hidden rounded-md border"
    style={{ height: 440 }}
  >
    <Sidebar collapsible="none" className="h-full border-r">
      <SidebarHeader>
        <div className="flex items-center justify-between p-2">
          <span className="text-sm font-medium">Kloudlite</span>
          <SidebarTrigger />
        </div>
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
