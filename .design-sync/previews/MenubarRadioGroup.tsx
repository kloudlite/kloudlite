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
        <MenubarLabel>Active environment</MenubarLabel>
        <MenubarSeparator />
        <MenubarRadioGroup value="production">
          <MenubarRadioItem value="development">development</MenubarRadioItem>
          <MenubarRadioItem value="staging">staging</MenubarRadioItem>
          <MenubarRadioItem value="production">production</MenubarRadioItem>
        </MenubarRadioGroup>
        </MenubarContent>
      </MenubarMenu>
      <Rest omit="environment" />
    </Menubar>
  </Frame>
)

export const WithActions = () => (
  <Frame>
    <Menubar defaultValue="environment">
      <MenubarMenu value="environment">
        <MenubarTrigger>Environment</MenubarTrigger>
        <MenubarContent className="w-64">
        <MenubarLabel>Active environment</MenubarLabel>
        <MenubarSeparator />
        <MenubarRadioGroup value="production">
          <MenubarRadioItem value="development">development</MenubarRadioItem>
          <MenubarRadioItem value="staging">staging</MenubarRadioItem>
          <MenubarRadioItem value="production">production</MenubarRadioItem>
        </MenubarRadioGroup>
        <MenubarSeparator />
        <MenubarItem>Copy kubeconfig</MenubarItem>
        </MenubarContent>
      </MenubarMenu>
      <Rest omit="environment" />
    </Menubar>
  </Frame>
)
