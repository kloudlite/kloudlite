import * as React from 'react'
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from '@kloudlite/ui'

export const SizedPanes = () => (
  <div className="w-96 border" style={{ height: 240 }}>
    <ResizablePanelGroup direction="horizontal">
      <ResizablePanel defaultSize="45%" minSize="25%" maxSize="60%">
        <div className="h-full bg-muted p-3 text-sm">
          <p className="font-medium">Namespaces</p>
          <p className="text-muted-foreground">kl-core</p>
          <p className="text-muted-foreground">kl-tenant</p>
        </div>
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel defaultSize="55%">
        <div className="h-full p-3 text-sm">
          <p className="font-medium">kl-tenant</p>
          <p className="text-muted-foreground">18 pods &middot; 4 services</p>
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  </div>
)

export const CollapsedPanel = () => (
  <div className="w-96 border" style={{ height: 240 }}>
    <ResizablePanelGroup direction="horizontal">
      <ResizablePanel defaultSize="15%" collapsible collapsedSize="15%" minSize="15%">
        <div className="h-full bg-muted p-3 text-center text-xs">kl</div>
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel defaultSize="85%">
        <div className="h-full p-3 text-sm">
          <p className="font-medium">Editor</p>
          <p className="text-muted-foreground">
            The rail on the left is collapsed to its minimum size.
          </p>
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  </div>
)

export const StackedDetails = () => (
  <div className="w-96 border" style={{ height: 240 }}>
    <ResizablePanelGroup direction="horizontal">
      <ResizablePanel defaultSize="50%" minSize="25%">
        <div className="h-full p-3 text-sm">
          <p className="font-medium">api-gateway-7d9f</p>
          <p className="text-muted-foreground">Running</p>
          <p className="text-muted-foreground">restarts 0</p>
        </div>
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel defaultSize="50%">
        <div className="h-full p-3 font-mono text-xs text-muted-foreground">
          <p>listening on :3000</p>
          <p>connected to mongo</p>
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  </div>
)
