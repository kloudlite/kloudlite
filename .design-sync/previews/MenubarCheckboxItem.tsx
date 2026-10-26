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

export const CheckedAndUnchecked = () => (
  <Frame>
    <Menubar defaultValue="view">
      <MenubarMenu value="view">
        <MenubarTrigger>View</MenubarTrigger>
        <MenubarContent className="w-64">
        <MenubarLabel>Workload filters</MenubarLabel>
        <MenubarSeparator />
        <MenubarCheckboxItem checked>Show system namespaces</MenubarCheckboxItem>
        <MenubarCheckboxItem>Show completed pods</MenubarCheckboxItem>
        <MenubarCheckboxItem checked>Group by namespace</MenubarCheckboxItem>
        </MenubarContent>
      </MenubarMenu>
      <Rest omit="view" />
    </Menubar>
  </Frame>
)

export const Disabled = () => (
  <Frame>
    <Menubar defaultValue="view">
      <MenubarMenu value="view">
        <MenubarTrigger>View</MenubarTrigger>
        <MenubarContent className="w-64">
        <MenubarCheckboxItem checked>Follow log stream</MenubarCheckboxItem>
        <MenubarCheckboxItem disabled>Show previous container</MenubarCheckboxItem>
        </MenubarContent>
      </MenubarMenu>
      <Rest omit="view" />
    </Menubar>
  </Frame>
)
