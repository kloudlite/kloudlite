import * as React from 'react'
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from '@kloudlite/ui'

export const WithGrip = () => (
  <div className="w-96 border" style={{ height: 220 }}>
    <ResizablePanelGroup direction="horizontal">
      <ResizablePanel defaultSize="40%">
        <div className="h-full p-3 text-sm">
          <p className="font-medium">Environments</p>
          <p className="text-muted-foreground">staging</p>
          <p className="text-muted-foreground">production</p>
        </div>
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel defaultSize="60%">
        <div className="h-full p-3 text-sm text-muted-foreground">
          Drag the grip to resize.
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  </div>
)

export const HairlineHandle = () => (
  <div className="w-96 border" style={{ height: 220 }}>
    <ResizablePanelGroup direction="horizontal">
      <ResizablePanel defaultSize="50%">
        <div className="h-full p-3 text-sm">
          <p className="font-medium">Spec</p>
          <p className="text-muted-foreground">replicas: 3</p>
        </div>
      </ResizablePanel>
      <ResizableHandle />
      <ResizablePanel defaultSize="50%">
        <div className="h-full p-3 text-sm">
          <p className="font-medium">Status</p>
          <p className="text-muted-foreground">readyReplicas: 3</p>
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  </div>
)

export const NarrowRail = () => (
  <div className="w-96 border" style={{ height: 220 }}>
    <ResizablePanelGroup direction="horizontal">
      <ResizablePanel defaultSize="30%" minSize="20%">
        <div className="h-full bg-muted p-3 text-sm">
          <p className="font-medium">Workspace</p>
        </div>
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel defaultSize="70%">
        <div className="h-full p-3 font-mono text-xs text-muted-foreground">
          <p>$ kl status</p>
          <p>connected &middot; staging</p>
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  </div>
)
