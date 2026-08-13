import * as React from 'react'
import {
  Menubar,
  MenubarCheckboxItem,
  MenubarContent,
  MenubarGroup,
  MenubarItem,
  MenubarLabel,
  MenubarMenu,
  MenubarPortal,
  MenubarRadioGroup,
  MenubarRadioItem,
  MenubarSeparator,
  MenubarShortcut,
  MenubarSub,
  MenubarSubContent,
  MenubarSubTrigger,
  MenubarTrigger,
} from '@kloudlite/ui'

// Menu content only paints while its menu is open, so every preview drives the
// Menubar with `defaultValue` pointing at the menu that is the subject here.
const Frame = ({ children }: { children: React.ReactNode }) => (
  <div className="w-96 h-96">{children}</div>
)

const MENUS: Array<[string, string]> = [
  ['file', 'File'],
  ['edit', 'Edit'],
  ['view', 'View'],
  ['help', 'Help'],
]

// The remaining top-level menus, minus whichever one this preview opens.
const Rest = ({ omit }: { omit?: string }) => (
  <>
    {MENUS.filter(([value]) => value !== omit).map(([value, label]) => (
      <MenubarMenu key={value} value={value}>
        <MenubarTrigger>{label}</MenubarTrigger>
      </MenubarMenu>
    ))}
  </>
)

export const Basic = () => (
  <Frame>
    <Menubar defaultValue="environment">
      <MenubarMenu value="environment">
        <MenubarTrigger>Environment</MenubarTrigger>
        <MenubarContent className="w-64">
        <MenubarItem>Restart deployment</MenubarItem>
        <MenubarSub defaultOpen>
          <MenubarSubTrigger>Switch cluster</MenubarSubTrigger>
          <MenubarSubContent forceMount className="w-56">
            <MenubarItem>kl-dev &middot; ap-south-1</MenubarItem>
            <MenubarItem>kl-staging &middot; eu-west-1</MenubarItem>
            <MenubarItem>kl-prod &middot; us-east-1</MenubarItem>
          </MenubarSubContent>
        </MenubarSub>
        <MenubarSeparator />
        <MenubarItem>Copy kubeconfig</MenubarItem>
        </MenubarContent>
      </MenubarMenu>
      <Rest omit="environment" />
    </Menubar>
  </Frame>
)

export const Inset = () => (
  <Frame>
    <Menubar defaultValue="view">
      <MenubarMenu value="view">
        <MenubarTrigger>View</MenubarTrigger>
        <MenubarContent className="w-64">
        <MenubarLabel>Layout</MenubarLabel>
        <MenubarItem inset>Toggle terminal</MenubarItem>
        <MenubarSub defaultOpen>
          <MenubarSubTrigger inset>Appearance</MenubarSubTrigger>
          <MenubarSubContent forceMount className="w-56">
            <MenubarItem>Light</MenubarItem>
            <MenubarItem>Dark</MenubarItem>
            <MenubarItem>Match system</MenubarItem>
          </MenubarSubContent>
        </MenubarSub>
        </MenubarContent>
      </MenubarMenu>
      <Rest omit="view" />
    </Menubar>
  </Frame>
)
