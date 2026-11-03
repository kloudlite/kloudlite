import * as React from 'react'
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@kloudlite/ui'

export const Basic = () => (
  <Card className="w-96">
    <CardHeader>
      <CardTitle>Production cluster</CardTitle>
      <CardDescription>
        eu-west-1 &middot; 12 nodes &middot; last synced 4 minutes ago
      </CardDescription>
    </CardHeader>
    <CardContent>
      <p className="text-sm text-muted-foreground">
        Workloads are reconciled from the environment spec. Changes to the spec roll out
        automatically once the cluster reports ready.
      </p>
    </CardContent>
    <CardFooter className="gap-2">
      <Button size="sm">Open</Button>
      <Button size="sm" variant="outline">
        Settings
      </Button>
    </CardFooter>
  </Card>
)

export const WithStatus = () => (
  <Card className="w-96">
    <CardHeader>
      <div className="flex items-start justify-between gap-3">
        <div className="space-y-1.5">
          <CardTitle>kloudlite-dev</CardTitle>
          <CardDescription>Installation &middot; Kloudlite Cloud</CardDescription>
        </div>
        <Badge>Active</Badge>
      </div>
    </CardHeader>
    <CardContent className="space-y-3 text-sm">
      <div className="flex justify-between">
        <span className="text-muted-foreground">Region</span>
        <span className="font-medium">ap-south-1</span>
      </div>
      <div className="flex justify-between">
        <span className="text-muted-foreground">Kubernetes</span>
        <span className="font-medium">v1.31.4</span>
      </div>
      <div className="flex justify-between">
        <span className="text-muted-foreground">Created</span>
        <span className="font-medium">6 Feb 2026</span>
      </div>
    </CardContent>
  </Card>
)

export const ContentOnly = () => (
  <Card className="w-96">
    <CardContent className="pt-6">
      <p className="text-sm">
        A card with no header or footer &mdash; useful for stat tiles and inline panels.
      </p>
    </CardContent>
  </Card>
)
