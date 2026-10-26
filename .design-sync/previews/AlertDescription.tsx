import * as React from 'react'
import { Alert, AlertDescription, AlertTitle } from '@kloudlite/ui'

export const InDefaultAlert = () => (
  <Alert className="w-96">
    <AlertTitle>Environment ready</AlertTitle>
    <AlertDescription>
      kloudlite-dev is reachable at dev.kloudlite.app. Your tunnel is connected and all
      6 deployments report healthy.
    </AlertDescription>
  </Alert>
)

export const InDestructiveAlert = () => (
  <Alert variant="destructive" className="w-96">
    <AlertTitle>Uninstall incomplete</AlertTitle>
    <AlertDescription>
      The cleanup job left 2 namespaces behind. Remove them manually before installing
      again on this cluster.
    </AlertDescription>
  </Alert>
)

export const WithParagraphs = () => (
  <Alert className="w-96">
    <AlertTitle>Kubeconfig rotated</AlertTitle>
    <AlertDescription className="space-y-2">
      <p>The service-account token for Production cluster was rotated an hour ago.</p>
      <p className="text-muted-foreground">
        Re-run kl cluster kubeconfig to refresh your local credentials.
      </p>
    </AlertDescription>
  </Alert>
)
