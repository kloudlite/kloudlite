import { useMemo, useState, useCallback } from 'react'
import {
  ReactFlow,
  ReactFlowProvider,
  Background,
  BackgroundVariant,
  Panel,
  useReactFlow,
  type Node,
  type Edge,
  type EdgeTypes,
  type NodeTypes,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { Plus, Minus, Maximize2, Grid3x3, Layers } from 'lucide-react'
import { cn } from '@/lib/utils'
import { ServiceNode } from './service-node'
import { WorkspaceNode } from './workspace-node'

export interface GraphPort {
  port: number
  targetPort: number
  interceptedBy?: string
}

export interface GraphVolume {
  name: string
  mountPath: string
  type: 'persistent' | 'config' | 'secret' | 'host'
}

export interface GraphService {
  id: string
  name: string
  dns: string
  type: 'ClusterIP' | 'LoadBalancer' | 'NodePort'
  ports: GraphPort[]
  volumes: GraphVolume[]
}

export interface GraphWorkspace {
  id: string
  name: string
  owner: string
  status: 'running' | 'stopped' | 'failed'
}

interface ServicesGraphProps {
  services: GraphService[]
  workspaces: GraphWorkspace[]
}

const nodeTypes: NodeTypes = {
  service: ServiceNode,
  workspace: WorkspaceNode,
}

const edgeTypes: EdgeTypes = {}

const SERVICE_NODE_W = 320
const WORKSPACE_NODE_W = 280
const COL_GAP = 280       // gap between services and workspaces columns
const ROW_GAP = 32        // consistent gap between cards
const HEADER_H = 64
const PORT_ROW_H = 36
const VOLUME_HEADER_H = 28
const VOLUME_ROW_H = 28
const WORKSPACE_H = 100
const MIN_ZOOM = 0.5
const MAX_ZOOM = 1.5

const SERVICE_X = 0
const WORKSPACE_X = SERVICE_NODE_W + COL_GAP

function getServiceHeight(svc: { ports: unknown[]; volumes: unknown[] }): number {
  return HEADER_H + svc.ports.length * PORT_ROW_H + (svc.volumes.length > 0 ? VOLUME_HEADER_H + svc.volumes.length * VOLUME_ROW_H : 0)
}

function GraphInner({ services, workspaces }: ServicesGraphProps) {
  const { zoomIn, zoomOut, fitView } = useReactFlow()
  const [snapToGrid, setSnapToGrid] = useState(true)
  const [showMinimap, setShowMinimap] = useState(false)

  const { nodes, edges } = useMemo(() => {
    // Stack services vertically on the left with consistent gap
    let cursorY = 0
    const serviceNodes: Node[] = services.map((svc) => {
      const interceptedPorts = svc.ports.filter((p) => p.interceptedBy)
      const node: Node = {
        id: `svc-${svc.id}`,
        type: 'service',
        position: { x: SERVICE_X, y: cursorY },
        data: {
          name: svc.name,
          dns: svc.dns,
          ports: svc.ports,
          volumes: svc.volumes,
          type: svc.type,
          interceptedCount: interceptedPorts.length,
          workspaceMap: Object.fromEntries(workspaces.map((w) => [w.id, w.name])),
        },
        draggable: true,
      }
      cursorY += getServiceHeight(svc) + ROW_GAP
      return node
    })

    const totalServiceHeight = cursorY - ROW_GAP
    const totalWsHeight = workspaces.length * WORKSPACE_H + (workspaces.length - 1) * ROW_GAP
    const wsStartY = Math.max(0, (totalServiceHeight - totalWsHeight) / 2)

    const wsNodes: Node[] = workspaces.map((ws, i) => {
      const interceptCount = services.reduce(
        (sum, s) => sum + s.ports.filter((p) => p.interceptedBy === ws.id).length,
        0
      )
      return {
        id: `ws-${ws.id}`,
        type: 'workspace',
        position: { x: WORKSPACE_X, y: wsStartY + i * (WORKSPACE_H + ROW_GAP) },
        data: {
          name: ws.name,
          owner: ws.owner,
          status: ws.status,
          interceptCount,
        },
        draggable: true,
      }
    })

    // One edge per intercepted port
    const interceptEdges: Edge[] = []
    for (const svc of services) {
      for (const p of svc.ports) {
        if (p.interceptedBy) {
          interceptEdges.push({
            id: `intercept-${svc.id}-${p.port}`,
            source: `svc-${svc.id}`,
            sourceHandle: `port-${p.port}`,
            target: `ws-${p.interceptedBy}`,
            type: 'smoothstep',
            animated: true,
            label: `:${p.port}`,
            labelStyle: { fontSize: 11, fontWeight: 600, fill: '#d97706', fontFamily: 'SF Mono, Menlo, monospace' },
            labelBgStyle: { fill: '#fef3c7', fillOpacity: 0.95 },
            labelBgPadding: [4, 8] as [number, number],
            labelBgBorderRadius: 6,
            style: { stroke: '#f59e0b', strokeWidth: 2, strokeDasharray: '6 4' },
          })
        }
      }
    }

    return { nodes: [...serviceNodes, ...wsNodes], edges: interceptEdges }
  }, [services, workspaces])

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      fitView
      fitViewOptions={{ padding: 0.1, maxZoom: 1, minZoom: MIN_ZOOM }}
      minZoom={MIN_ZOOM}
      maxZoom={MAX_ZOOM}
      defaultViewport={{ x: 0, y: 0, zoom: 1 }}
      snapToGrid={snapToGrid}
      snapGrid={[16, 16]}
      proOptions={{ hideAttribution: true }}
    >
      <Background variant={BackgroundVariant.Dots} gap={16} size={1} />

      {/* Custom toolbar */}
      <Panel position="bottom-right" className="!m-4 flex flex-col gap-1.5">
        <ToolButton onClick={() => setSnapToGrid(!snapToGrid)} active={snapToGrid} tooltip="Snap to grid">
          <Grid3x3 className="h-4 w-4" />
        </ToolButton>
        <div className="my-1 h-px bg-border/50" />
        <ToolButton onClick={() => zoomIn({ duration: 200 })} tooltip="Zoom in">
          <Plus className="h-4 w-4" />
        </ToolButton>
        <ToolButton onClick={() => zoomOut({ duration: 200 })} tooltip="Zoom out">
          <Minus className="h-4 w-4" />
        </ToolButton>
        <ToolButton onClick={() => fitView({ duration: 300, padding: 0.1, maxZoom: 1, minZoom: MIN_ZOOM })} tooltip="Fit view">
          <Maximize2 className="h-4 w-4" />
        </ToolButton>
        <div className="my-1 h-px bg-border/50" />
        <ToolButton onClick={() => setShowMinimap(!showMinimap)} active={showMinimap} tooltip="Toggle minimap">
          <Layers className="h-4 w-4" />
        </ToolButton>
      </Panel>
    </ReactFlow>
  )
}

function ToolButton({ children, onClick, active, tooltip }: { children: React.ReactNode; onClick: () => void; active?: boolean; tooltip?: string }) {
  return (
    <button
      title={tooltip}
      className={cn(
        'flex h-8 w-8 items-center justify-center rounded-lg border bg-background/90 backdrop-blur-sm transition-all',
        active
          ? 'border-primary/40 bg-primary/10 text-primary shadow-sm'
          : 'border-border/60 text-muted-foreground hover:border-border hover:bg-accent hover:text-foreground'
      )}
      onClick={onClick}
    >
      {children}
    </button>
  )
}

export function ServicesGraph(props: ServicesGraphProps) {
  return (
    <div className="h-full w-full">
      <ReactFlowProvider>
        <GraphInner {...props} />
      </ReactFlowProvider>
    </div>
  )
}
