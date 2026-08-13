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
      <CardTitle>Cluster overview</CardTitle>
      <CardDescription>ap-south-1</CardDescription>
    </CardHeader>
    <CardContent className="text-sm text-muted-foreground">
      Workloads are reconciled automatically whenever the environment spec changes.
    </CardContent>
  </Card>
)

export const KeyValueRows = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle>Installation details</CardTitle>
    </CardHeader>
    <CardContent className="space-y-2 text-sm">
      <div className="flex justify-between">
        <span className="text-muted-foreground">Region</span>
        <span className="font-medium">eu-west-1</span>
      </div>
      <div className="flex justify-between">
        <span className="text-muted-foreground">Kubernetes</span>
        <span className="font-medium">v1.31.4</span>
      </div>
      <div className="flex justify-between">
        <span className="text-muted-foreground">Node pools</span>
        <span className="font-medium">3</span>
      </div>
    </CardContent>
  </Card>
)

export const ContentOnly = () => (
  <Card className="w-80">
    <CardContent className="p-6">
      <div className="text-sm text-muted-foreground">Active intercepts</div>
      <div className="text-2xl font-semibold">7</div>
    </CardContent>
  </Card>
)
