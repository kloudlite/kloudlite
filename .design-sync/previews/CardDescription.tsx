import * as React from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@kloudlite/ui'

export const Basic = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle>api-gateway</CardTitle>
      <CardDescription>Deployment &middot; namespace kl-platform</CardDescription>
    </CardHeader>
    <CardContent className="text-sm text-muted-foreground">
      Image tag ghcr.io/kloudlite/api-gateway:v1.9.2
    </CardContent>
  </Card>
)

export const MultiLine = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle>Bring your own cluster</CardTitle>
      <CardDescription>
        Attach an existing Kubernetes cluster to Kloudlite. We install the operator stack
        into the kl-core namespace and never touch your other workloads.
      </CardDescription>
    </CardHeader>
  </Card>
)

export const AsMetadataLine = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle>workmachine-a3f9</CardTitle>
      <CardDescription>
        8 vCPU &middot; 32 GiB &middot; eu-west-1b &middot; up 3 days
      </CardDescription>
    </CardHeader>
    <CardContent className="text-sm text-muted-foreground">
      Connected through the tunnel at 10.42.0.17.
    </CardContent>
  </Card>
)
