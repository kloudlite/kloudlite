import * as React from 'react'
import { Badge, Button, Label, Switch, Tabs, TabsContent, TabsList, TabsTrigger } from '@kloudlite/ui'

export const Basic = () => (
  <Tabs defaultValue="overview" className="w-96">
    <TabsList>
      <TabsTrigger value="overview">Overview</TabsTrigger>
      <TabsTrigger value="logs">Logs</TabsTrigger>
      <TabsTrigger value="settings">Settings</TabsTrigger>
    </TabsList>
    <TabsContent value="overview">
      <div className="space-y-2 border p-4 text-sm">
        <div className="flex justify-between">
          <span className="text-muted-foreground">Namespace</span>
          <span className="font-medium">kloudlite-dev</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Replicas</span>
          <span className="font-medium tabular-nums">3 / 3</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Image</span>
          <span className="font-medium">api-gateway:v2.14.0</span>
        </div>
      </div>
    </TabsContent>
    <TabsContent value="logs">
      <div className="border p-4 text-xs text-muted-foreground">
        No log lines in the selected window.
      </div>
    </TabsContent>
    <TabsContent value="settings">
      <div className="border p-4 text-sm">Settings for api-gateway.</div>
    </TabsContent>
  </Tabs>
)

export const LogsSelected = () => (
  <Tabs defaultValue="logs" className="w-96">
    <TabsList>
      <TabsTrigger value="overview">Overview</TabsTrigger>
      <TabsTrigger value="logs">Logs</TabsTrigger>
      <TabsTrigger value="settings">Settings</TabsTrigger>
    </TabsList>
    <TabsContent value="overview">
      <div className="border p-4 text-sm">Overview of api-gateway.</div>
    </TabsContent>
    <TabsContent value="logs">
      <div className="space-y-1 border p-4 font-mono text-xs text-muted-foreground">
        <div>12:04:11 listening on :8080</div>
        <div>12:04:12 connected to postgres</div>
        <div>12:04:19 GET /v1/environments 200 14ms</div>
        <div>12:05:02 GET /v1/deployments 200 31ms</div>
      </div>
    </TabsContent>
    <TabsContent value="settings">
      <div className="border p-4 text-sm">Settings for api-gateway.</div>
    </TabsContent>
  </Tabs>
)

export const SettingsPanel = () => (
  <Tabs defaultValue="settings" className="w-96">
    <TabsList>
      <TabsTrigger value="overview">Overview</TabsTrigger>
      <TabsTrigger value="settings">Settings</TabsTrigger>
      <TabsTrigger value="danger" disabled>
        Danger
      </TabsTrigger>
    </TabsList>
    <TabsContent value="overview">
      <div className="border p-4 text-sm">Overview of kloudlite-prod.</div>
    </TabsContent>
    <TabsContent value="settings">
      <div className="space-y-4 border p-4">
        <div className="flex items-center justify-between gap-4">
          <Label>Auto-scale node pool</Label>
          <Switch defaultChecked />
        </div>
        <div className="flex items-center justify-between gap-4">
          <Label>Public ingress</Label>
          <Switch />
        </div>
        <Button size="sm">Save changes</Button>
      </div>
    </TabsContent>
    <TabsContent value="danger">
      <div className="border p-4 text-sm">Destructive actions.</div>
    </TabsContent>
  </Tabs>
)
