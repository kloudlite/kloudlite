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
} from '@kloudlite/ui'

export const BehindContent = () => (
  <AlertDialog open>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Delete environment staging?</AlertDialogTitle>
        <AlertDialogDescription>
          The dimmed backdrop behind this dialog is the overlay. It blocks interaction
          with the console until a choice is made.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction>Delete environment</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
)

export const WithPageContentBehind = () => (
  <div className="w-96 space-y-2 p-6">
    <p className="text-sm font-medium">Installations</p>
    <p className="text-sm text-muted-foreground">
      kloudlite-dev &middot; ap-south-1 &middot; 6 deployments
    </p>
    <p className="text-sm text-muted-foreground">
      Production cluster &middot; eu-west-1 &middot; 12 nodes
    </p>
    <AlertDialog open>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Confirm removal</AlertDialogTitle>
          <AlertDialogDescription>
            The list behind is dimmed by the overlay.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction>Remove</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
)
