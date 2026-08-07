import * as React from 'react'
import {
  Badge,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@kloudlite/ui'

export const Basic = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle>Production cluster</CardTitle>
      <CardDescription>ap-south-1 &middot; Kubernetes v1.31.4</CardDescription>
    </CardHeader>
    <CardContent className="text-sm text-muted-foreground">
      18 deployments across 5 namespaces.
    </CardContent>
  </Card>
)

export const LargeTitle = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle className="text-2xl">1,284</CardTitle>
      <CardDescription>Intercepts served this week</CardDescription>
    </CardHeader>
  </Card>
)

export const TitleWithBadge = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle className="flex items-center gap-2">
        kloudlite-dev
        <Badge>Active</Badge>
      </CardTitle>
      <CardDescription>Installation on Kloudlite Cloud</CardDescription>
    </CardHeader>
    <CardContent className="text-sm text-muted-foreground">
      Owned by platform@kloudlite.io
    </CardContent>
  </Card>
)
