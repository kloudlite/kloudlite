import * as React from 'react'
import { Alert, AlertDescription, AlertTitle } from '@kloudlite/ui'

export const Default = () => (
  <Alert className="w-96">
    <AlertTitle>Cluster upgrade scheduled</AlertTitle>
    <AlertDescription>
      Production cluster in eu-west-1 will be upgraded to Kubernetes v1.31.4 on 14 Aug.
      Workloads are drained node by node.
    </AlertDescription>
  </Alert>
)

export const Destructive = () => (
  <Alert variant="destructive" className="w-96">
    <AlertTitle>Installation failed</AlertTitle>
    <AlertDescription>
      The kloudlite agent could not reach the cluster API server. Check that the
      kubeconfig is valid and the endpoint is publicly reachable.
    </AlertDescription>
  </Alert>
)

export const TitleOnly = () => (
  <Alert className="w-96">
    <AlertTitle>Tunnel reconnected to ap-south-1</AlertTitle>
  </Alert>
)

export const DescriptionOnly = () => (
  <Alert className="w-96">
    <AlertDescription>
      Intercepts for payments-api are routed to your workmachine until you stop the
      session.
    </AlertDescription>
  </Alert>
)
