import * as React from 'react'
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@kloudlite/ui'

export const Basic = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle>staging-eu</CardTitle>
      <CardDescription>Environment &middot; eu-west-1 &middot; 4 workspaces</CardDescription>
    </CardHeader>
    <CardContent className="text-sm text-muted-foreground">
      Reconciled from the environment spec 12 minutes ago.
    </CardContent>
  </Card>
)

export const WithTrailingAction = () => (
  <Card className="w-80">
    <CardHeader className="flex-row items-start justify-between gap-4 space-y-0">
      <div className="space-y-2">
        <CardTitle>node-pool-spot</CardTitle>
        <CardDescription>6 nodes &middot; c6a.2xlarge</CardDescription>
      </div>
      <Badge variant="secondary">Scaling</Badge>
    </CardHeader>
    <CardContent className="text-sm text-muted-foreground">
      Two nodes are joining the cluster and will accept workloads shortly.
    </CardContent>
  </Card>
)

export const HeaderOnly = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle>Tunnel credentials</CardTitle>
      <CardDescription>
        Rotate the wireguard keys used by every workmachine in this installation.
      </CardDescription>
      <Button size="sm" variant="outline" className="w-40">
        Rotate keys
      </Button>
    </CardHeader>
  </Card>
)
