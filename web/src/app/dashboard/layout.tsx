"use client";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupAction,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
} from "@/components/ui/sidebar";
import {
  ArrowRight,
  ArrowUpRight,
  Boxes,
  ChevronDown,
  ChevronsDownUp,
  ChevronsUpDown,
  Package,
  Settings,
  SquareTerminal,
  Users,
} from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

export default function Dashboard({ children }: { children: React.ReactNode }) {

  const pathName = usePathname();
  

  return (
    <SidebarProvider>
      <Sidebar className="h-full" side="left" variant="sidebar">
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <SidebarMenuButton className="p-2 h-12">
                    <div className="flex flex-col">
                      <span className="font-medium">Smartcomms Team</span>
                      <span className="text-muted-foreground text-xs ml-1">
                        #smartcomms
                      </span>
                    </div>
                    <ChevronsUpDown className="ml-auto" />
                  </SidebarMenuButton>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  side="right"
                  className="w-[--radix-popper-anchor-width]"
                >
                  <DropdownMenuItem>
                    <span>Acme Inc</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <span>Acme Corp.</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link href={"/teams"}>
                      <span className="flex items-center w-full justify-between text-xs font-semibold text-muted-foreground">
                        View All
                        <ArrowUpRight className="ml-auto" />
                      </span>
                    </Link>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarGroupLabel>Workloads</SidebarGroupLabel>
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton asChild  isActive={pathName.startsWith("/dashboard/environments")}>
                    <Link href="/dashboard/environments">
                      <Boxes strokeWidth={2} />
                      Environments
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
                <SidebarMenuItem>
                  <SidebarMenuButton asChild  isActive={pathName.startsWith("/dashboard/services")}>
                    <Link href="/dashboard/services">
                      <Package />
                      Common Services
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
                <SidebarMenuItem>
                  <SidebarMenuButton asChild isActive={pathName.startsWith("/dashboard/workspaces")}>
                    <Link href="/dashboard/workspaces">
                      <SquareTerminal />
                      Workspaces
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarGroupLabel>Management</SidebarGroupLabel>
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton asChild isActive={pathName.startsWith("/dashboard/settings")}>
                    <Link href="/dashboard/settings">
                      <Settings />  
                      Settings
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton asChild isActive={pathName.startsWith("/dashboard/users")}>
                    <Link href="/dashboard/users">
                      <Users />
                      Users
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter />
      </Sidebar>
      <SidebarInset>
        <main>
          {children}
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}
