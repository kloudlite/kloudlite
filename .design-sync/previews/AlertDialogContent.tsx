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

export const Open = () => (
  <AlertDialog open>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Delete Production cluster?</AlertDialogTitle>
        <AlertDialogDescription>
          12 nodes in eu-west-1 will be terminated and every environment scheduled on
          them will stop.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction>Delete cluster</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
)

export const WithExtraBody = () => (
  <AlertDialog open>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Roll back payments-api?</AlertDialogTitle>
        <AlertDialogDescription>
          Traffic returns to the previously deployed image digest.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <div className="space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-muted-foreground">Current</span>
          <span className="font-medium">payments-api:2026.8.3</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Rolling back to</span>
          <span className="font-medium">payments-api:2026.7.28</span>
        </div>
      </div>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction>Roll back</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
)

export const ShortConfirm = () => (
  <AlertDialog open>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Stop intercept?</AlertDialogTitle>
        <AlertDialogDescription>
          Traffic goes back to the in-cluster deployment.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Keep running</AlertDialogCancel>
        <AlertDialogAction>Stop</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
)
