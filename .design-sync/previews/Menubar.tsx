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

export const FileMenuOpen = () => (
  <Frame>
    <Menubar defaultValue="file">
      <MenubarMenu value="file">
        <MenubarTrigger>File</MenubarTrigger>
        <MenubarContent className="w-64">
        <MenubarItem>
          New environment
          <MenubarShortcut>&#8984;N</MenubarShortcut>
        </MenubarItem>
        <MenubarItem>
          Open workspace&hellip;
          <MenubarShortcut>&#8984;O</MenubarShortcut>
        </MenubarItem>
        <MenubarSeparator />
        <MenubarItem>
          Copy kubeconfig
          <MenubarShortcut>&#8984;&#8679;K</MenubarShortcut>
        </MenubarItem>
        </MenubarContent>
      </MenubarMenu>
      <Rest omit="file" />
    </Menubar>
  </Frame>
)

export const EnvironmentMenuOpen = () => (
  <Frame>
    <Menubar defaultValue="environment">
      <MenubarMenu value="environment">
        <MenubarTrigger>Environment</MenubarTrigger>
        <MenubarContent className="w-64">
        <MenubarItem>Intercept service&hellip;</MenubarItem>
        <MenubarItem>Restart deployment</MenubarItem>
        <MenubarItem>Stream pod logs</MenubarItem>
        <MenubarSeparator />
        <MenubarItem>Copy kubeconfig</MenubarItem>
        </MenubarContent>
      </MenubarMenu>
      <Rest omit="environment" />
    </Menubar>
  </Frame>
)

export const Closed = () => (
  <Frame>
    <Menubar>
      <MenubarMenu value="file">
        <MenubarTrigger>File</MenubarTrigger>
      </MenubarMenu>
      <Rest omit="file" />
    </Menubar>
  </Frame>
)
