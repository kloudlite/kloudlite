import * as React from 'react'
import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuIndicator,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
} from '@kloudlite/ui'

const linkClass =
  'block rounded-md p-3 hover:bg-accent'
const topLinkClass =
  'flex h-9 items-center rounded-md px-4 py-2 text-sm font-medium hover:bg-accent'

const Panel = ({ children }: { children: React.ReactNode }) => (
  <div className="relative h-80 w-full p-6">{children}</div>
)

const PlatformPanel = () => (
  <ul className="grid w-96 grid-cols-2 gap-3 p-4">
    <li>
      <NavigationMenuLink className={linkClass}>
        <div className="text-sm font-medium leading-none">Environments</div>
        <p className="text-sm text-muted-foreground">
          Ephemeral namespaces that mirror production on your own cluster.
        </p>
      </NavigationMenuLink>
    </li>
    <li>
      <NavigationMenuLink className={linkClass}>
        <div className="text-sm font-medium leading-none">Workspaces</div>
        <p className="text-sm text-muted-foreground">
          Remote dev containers with your editor attached over the tunnel.
        </p>
      </NavigationMenuLink>
    </li>
    <li>
      <NavigationMenuLink className={linkClass}>
        <div className="text-sm font-medium leading-none">Clusters</div>
        <p className="text-sm text-muted-foreground">
          Attach EKS, GKE or bring your own kubeconfig.
        </p>
      </NavigationMenuLink>
    </li>
    <li>
      <NavigationMenuLink className={linkClass}>
        <div className="text-sm font-medium leading-none">Installations</div>
        <p className="text-sm text-muted-foreground">
          Track agent rollout and reconciliation health per cluster.
        </p>
      </NavigationMenuLink>
    </li>
  </ul>
)

const ResourcesPanel = () => (
  <ul className="grid w-64 gap-3 p-4">
    <li>
      <NavigationMenuLink className={linkClass}>
        <div className="text-sm font-medium leading-none">Documentation</div>
        <p className="text-sm text-muted-foreground">Guides and CLI reference.</p>
      </NavigationMenuLink>
    </li>
    <li>
      <NavigationMenuLink className={linkClass}>
        <div className="text-sm font-medium leading-none">Changelog</div>
        <p className="text-sm text-muted-foreground">What shipped this week.</p>
      </NavigationMenuLink>
    </li>
  </ul>
)

export const PopoverSurface = () => (
  <Panel>
  <NavigationMenu defaultValue="platform">
    <NavigationMenuList>
      <NavigationMenuItem value="platform">
        <NavigationMenuTrigger>Platform</NavigationMenuTrigger>
        <NavigationMenuContent>
          <PlatformPanel />
        </NavigationMenuContent>
      </NavigationMenuItem>
      <NavigationMenuItem value="resources">
        <NavigationMenuTrigger>Resources</NavigationMenuTrigger>
        <NavigationMenuContent>
          <ResourcesPanel />
        </NavigationMenuContent>
      </NavigationMenuItem>
      <NavigationMenuItem>
        <NavigationMenuLink className={topLinkClass}>Team</NavigationMenuLink>
      </NavigationMenuItem>
      <NavigationMenuItem>
        <NavigationMenuLink className={topLinkClass}>Settings</NavigationMenuLink>
      </NavigationMenuItem>
    </NavigationMenuList>
  </NavigationMenu>
  </Panel>
)

export const NarrowSurface = () => (
  <Panel>
  <NavigationMenu defaultValue="resources">
    <NavigationMenuList>
      <NavigationMenuItem value="platform">
        <NavigationMenuTrigger>Platform</NavigationMenuTrigger>
        <NavigationMenuContent>
          <PlatformPanel />
        </NavigationMenuContent>
      </NavigationMenuItem>
      <NavigationMenuItem value="resources">
        <NavigationMenuTrigger>Resources</NavigationMenuTrigger>
        <NavigationMenuContent>
          <ResourcesPanel />
        </NavigationMenuContent>
      </NavigationMenuItem>
      <NavigationMenuItem>
        <NavigationMenuLink className={topLinkClass}>Team</NavigationMenuLink>
      </NavigationMenuItem>
      <NavigationMenuItem>
        <NavigationMenuLink className={topLinkClass}>Settings</NavigationMenuLink>
      </NavigationMenuItem>
    </NavigationMenuList>
  </NavigationMenu>
  </Panel>
)
