import * as React from 'react'
import { Alert, AlertDescription, AlertTitle } from '@kloudlite/ui'

export const InDefaultAlert = () => (
  <Alert className="w-96">
    <AlertTitle>Node pool scaled up</AlertTitle>
    <AlertDescription>
      spot-a grew from 2 to 4 nodes to schedule pending workmachines.
    </AlertDescription>
  </Alert>
)

export const InDestructiveAlert = () => (
  <Alert variant="destructive" className="w-96">
    <AlertTitle>Credits exhausted</AlertTitle>
    <AlertDescription>
      Environments in kloudlite-dev are suspended until the account is topped up.
    </AlertDescription>
  </Alert>
)

export const LongTitle = () => (
  <Alert className="w-96">
    <AlertTitle>
      Three deployments in namespace kloudlite-dev are waiting on an image pull secret
    </AlertTitle>
    <AlertDescription>
      Add registry credentials to the environment to unblock the rollout.
    </AlertDescription>
  </Alert>
)
