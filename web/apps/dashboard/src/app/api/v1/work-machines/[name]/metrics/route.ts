import { NextRequest, NextResponse } from 'next/server'

const KUBE_API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

interface K8sMetricsResponse {
  metadata: {
    name: string
    creationTimestamp: string
  }
  timestamp: string
  window: string
  usage: {
    cpu: string
    memory: string
  }
}

interface K8sNodeResponse {
  metadata: {
    name: string
  }
  status: {
    capacity: {
      cpu: string
      memory: string
    }
    allocatable: {
      cpu: string
      memory: string
    }
  }
}

// Parse Kubernetes CPU format (e.g., "100m" = 0.1 cores, "2" = 2 cores)
function parseCPU(cpuString: string): number {
  if (cpuString.endsWith('n')) {
    return parseInt(cpuString.slice(0, -1)) / 1_000_000
  }
  if (cpuString.endsWith('u')) {
    return parseInt(cpuString.slice(0, -1)) / 1_000
  }
  if (cpuString.endsWith('m')) {
    return parseInt(cpuString.slice(0, -1))
  }
  return parseFloat(cpuString) * 1000 // Convert cores to millicores
}

// Parse Kubernetes memory format (e.g., "1024Ki", "1Mi", "1Gi")
function parseMemory(memoryString: string): number {
  const units: Record<string, number> = {
    Ki: 1024,
    Mi: 1024 * 1024,
    Gi: 1024 * 1024 * 1024,
    Ti: 1024 * 1024 * 1024 * 1024,
    K: 1000,
    M: 1000 * 1000,
    G: 1000 * 1000 * 1000,
    T: 1000 * 1000 * 1000 * 1000,
  }

  for (const [unit, multiplier] of Object.entries(units)) {
    if (memoryString.endsWith(unit)) {
      return parseInt(memoryString.slice(0, -unit.length)) * multiplier
    }
  }

  return parseInt(memoryString)
}

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ name: string }> }
) {
  try {
    const { name } = await params

    // Fetch node metrics from Kubernetes Metrics API
    const metricsResponse = await fetch(
      `${KUBE_API_URL}/apis/metrics.k8s.io/v1beta1/nodes/${name}`,
      {
        headers: {
          'Accept': 'application/json',
        },
        cache: 'no-store',
      }
    )

    if (!metricsResponse.ok) {
      return NextResponse.json(
        { error: `Failed to fetch node metrics: ${metricsResponse.status}` },
        { status: metricsResponse.status }
      )
    }

    const metricsData: K8sMetricsResponse = await metricsResponse.json()

    // Fetch node details for capacity and allocatable
    const nodeResponse = await fetch(
      `${KUBE_API_URL}/api/v1/nodes/${name}`,
      {
        headers: {
          'Accept': 'application/json',
        },
        cache: 'no-store',
      }
    )

    if (!nodeResponse.ok) {
      return NextResponse.json(
        { error: `Failed to fetch node details: ${nodeResponse.status}` },
        { status: nodeResponse.status }
      )
    }

    const nodeData: K8sNodeResponse = await nodeResponse.json()

    // Parse CPU values
    const cpuUsage = parseCPU(metricsData.usage.cpu)
    const cpuCapacity = parseCPU(nodeData.status.capacity.cpu)
    const cpuAllocatable = parseCPU(nodeData.status.allocatable.cpu)

    // Parse memory values
    const memoryUsage = parseMemory(metricsData.usage.memory)
    const memoryCapacity = parseMemory(nodeData.status.capacity.memory)
    const memoryAllocatable = parseMemory(nodeData.status.allocatable.memory)

    const response = {
      cpu: {
        usage: cpuUsage,
        capacity: cpuCapacity,
        allocatable: cpuAllocatable,
      },
      memory: {
        usage: memoryUsage,
        capacity: memoryCapacity,
        allocatable: memoryAllocatable,
      },
      timestamp: metricsData.timestamp,
    }

    return NextResponse.json(response)
  } catch (error) {
    console.error('Error fetching node metrics:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
