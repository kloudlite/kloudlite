import * as React from 'react'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
  Button,
} from '@kloudlite/ui'

const Body = () => (
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Revoke access for asha@kloudlite.io?</AlertDialogTitle>
      <AlertDialogDescription>
        She loses access to all environments in this installation immediately.
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <AlertDialogCancel>Cancel</AlertDialogCancel>
      <AlertDialogAction>Revoke access</AlertDialogAction>
    </AlertDialogFooter>
  </AlertDialogContent>
)

export const DestructiveButton = () => (
  <AlertDialog>
    <AlertDialogTrigger asChild>
      <Button variant="destructive">Revoke access</Button>
    </AlertDialogTrigger>
    <Body />
  </AlertDialog>
)

export const OutlineButton = () => (
  <AlertDialog>
    <AlertDialogTrigger asChild>
      <Button variant="outline" size="sm">
        Reset kubeconfig
      </Button>
    </AlertDialogTrigger>
    <Body />
  </AlertDialog>
)

export const LinkTrigger = () => (
  <AlertDialog>
    <AlertDialogTrigger asChild>
      <Button variant="link">Delete this installation</Button>
    </AlertDialogTrigger>
    <Body />
  </AlertDialog>
)
