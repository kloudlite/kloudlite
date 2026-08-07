import * as React from 'react'
import { Kbd, KbdGroup } from '@kloudlite/ui'

export const TwoKeyCombo = () => (
  <KbdGroup>
    <Kbd>⌘</Kbd>
    <Kbd>K</Kbd>
  </KbdGroup>
)

export const ThreeKeyCombo = () => (
  <KbdGroup>
    <Kbd>⌘</Kbd>
    <Kbd>⇧</Kbd>
    <Kbd>P</Kbd>
  </KbdGroup>
)

export const InlineWithText = () => (
  <p className="w-96 text-sm text-muted-foreground">
    Press{' '}
    <KbdGroup>
      <Kbd>⌘</Kbd>
      <Kbd>K</Kbd>
    </KbdGroup>{' '}
    to search environments, or{' '}
    <KbdGroup>
      <Kbd>Ctrl</Kbd>
      <Kbd>`</Kbd>
    </KbdGroup>{' '}
    to open a terminal in the workspace.
  </p>
)

export const StackedGroups = () => (
  <div className="flex flex-col gap-2">
    <KbdGroup>
      <Kbd>⌘</Kbd>
      <Kbd>K</Kbd>
    </KbdGroup>
    <KbdGroup>
      <Kbd>Ctrl</Kbd>
      <Kbd>C</Kbd>
    </KbdGroup>
    <KbdGroup>
      <Kbd>⌥</Kbd>
      <Kbd>↑</Kbd>
    </KbdGroup>
  </div>
)
