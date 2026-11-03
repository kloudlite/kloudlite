import * as React from 'react'
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@kloudlite/ui'

export const Basic = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle>staging-eu</CardTitle>
      <CardDescription>Environment &middot; eu-west-1</CardDescription>
    </CardHeader>
    <CardContent className="text-sm text-muted-foreground">
      4 workspaces are attached to this environment.
    </CardContent>
    <CardFooter className="gap-2">
      <Button size="sm">Open</Button>
      <Button size="sm" variant="outline">
        Settings
      </Button>
    </CardFooter>
  </Card>
)

export const SpaceBetween = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle>Uninstall cluster</CardTitle>
      <CardDescription>Removes the Kloudlite operator stack.</CardDescription>
    </CardHeader>
    <CardFooter className="justify-between">
      <span className="text-sm text-muted-foreground">Irreversible</span>
      <Button size="sm" variant="destructive">
        Uninstall
      </Button>
    </CardFooter>
  </Card>
)

export const FullWidthAction = () => (
  <Card className="w-80">
    <CardHeader>
      <CardTitle>Connect a cluster</CardTitle>
      <CardDescription>Run the install command on your own Kubernetes.</CardDescription>
    </CardHeader>
    <CardFooter>
      <Button className="w-full">Generate install command</Button>
    </CardFooter>
  </Card>
)
