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

export const Open = () => (
  <AlertDialog open>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Delete environment kloudlite-dev?</AlertDialogTitle>
        <AlertDialogDescription>
          This removes the namespace, all 6 deployments and the persistent volumes
          attached to it. This action cannot be undone.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction>Delete environment</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
)

export const ClosedWithTrigger = () => (
  <AlertDialog>
    <AlertDialogTrigger asChild>
      <Button variant="destructive">Uninstall from cluster</Button>
    </AlertDialogTrigger>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Uninstall Kloudlite?</AlertDialogTitle>
        <AlertDialogDescription>
          The agent and all managed namespaces are removed from Production cluster.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Keep installed</AlertDialogCancel>
        <AlertDialogAction>Uninstall</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
)

export const OpenDestructiveAction = () => (
  <AlertDialog open>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Remove node pool spot-a?</AlertDialogTitle>
        <AlertDialogDescription>
          4 nodes in ap-south-1 will be drained. Pods without another schedulable pool
          will stay pending.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction className="bg-destructive text-destructive-foreground">
          Remove pool
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
)
