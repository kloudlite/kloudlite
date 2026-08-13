import * as React from 'react'
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from '@kloudlite/ui'

export const WorkspaceShell = () => (
  <div className="w-96 border" style={{ height: 260 }}>
    <ResizablePanelGroup direction="horizontal">
      <ResizablePanel defaultSize="35%" minSize="20%">
        <div className="h-full space-y-2 p-3 text-sm">
          <p className="font-medium">kloudlite-api</p>
          <p className="text-muted-foreground">cmd/server</p>
          <p className="text-muted-foreground">pkg/reconciler</p>
          <p className="text-muted-foreground">charts/agent</p>
        </div>
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel defaultSize="65%">
        <div className="h-full p-3 font-mono text-xs leading-relaxed">
          <p>func main() {'{'}</p>
          <p>&nbsp;&nbsp;app.Run(ctx)</p>
          <p>{'}'}</p>
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  </div>
)

export const DetailsAndLogs = () => (
  <div className="w-96 border" style={{ height: 260 }}>
    <ResizablePanelGroup direction="horizontal">
      <ResizablePanel defaultSize="45%">
        <div className="h-full p-3 text-sm">
          <p className="font-medium">api-gateway</p>
          <p className="text-muted-foreground">3/3 ready</p>
          <p className="text-muted-foreground">v1.2.9</p>
        </div>
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel defaultSize="55%">
        <div className="h-full p-3 font-mono text-xs text-muted-foreground">
          <p>reconciled</p>
          <p>rollout complete</p>
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  </div>
)

export const ThreePanes = () => (
  <div className="w-96 border" style={{ height: 220 }}>
    <ResizablePanelGroup direction="horizontal">
      <ResizablePanel defaultSize="25%">
        <div className="h-full p-3 text-xs">Envs</div>
      </ResizablePanel>
      <ResizableHandle />
      <ResizablePanel defaultSize="45%">
        <div className="h-full p-3 text-xs">Workspaces</div>
      </ResizablePanel>
      <ResizableHandle />
      <ResizablePanel defaultSize="30%">
        <div className="h-full p-3 text-xs">Details</div>
      </ResizablePanel>
    </ResizablePanelGroup>
  </div>
)
