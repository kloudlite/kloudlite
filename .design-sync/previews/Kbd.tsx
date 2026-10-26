import * as React from 'react'
import { Kbd, KbdGroup } from '@kloudlite/ui'

export const SingleKey = () => <Kbd>K</Kbd>

export const CommandPalette = () => (
  <div className="flex items-center gap-2 text-sm">
    <span className="text-muted-foreground">Open command palette</span>
    <KbdGroup>
      <Kbd>⌘</Kbd>
      <Kbd>K</Kbd>
    </KbdGroup>
  </div>
)

export const ShortcutList = () => (
  <div className="w-80 space-y-2 text-sm">
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">Switch environment</span>
      <KbdGroup>
        <Kbd>⌘</Kbd>
        <Kbd>⇧</Kbd>
        <Kbd>E</Kbd>
      </KbdGroup>
    </div>
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">Copy kubeconfig</span>
      <KbdGroup>
        <Kbd>Ctrl</Kbd>
        <Kbd>C</Kbd>
      </KbdGroup>
    </div>
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">Reload workspace</span>
      <KbdGroup>
        <Kbd>⌘</Kbd>
        <Kbd>R</Kbd>
      </KbdGroup>
    </div>
  </div>
)

export const ArrowKeys = () => (
  <div className="flex items-center gap-2 text-sm">
    <span className="text-muted-foreground">Navigate results</span>
    <KbdGroup>
      <Kbd>↑</Kbd>
      <Kbd>↓</Kbd>
    </KbdGroup>
    <span className="text-muted-foreground">select</span>
    <Kbd>↵</Kbd>
  </div>
)
