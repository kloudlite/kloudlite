export interface ComposeServicePort {
  port: number
  targetPort: number
  protocol: string
  interceptedBy?: string
}

export interface ComposeServiceVolume {
  name: string
  mountPath: string
  type: 'persistent' | 'config' | 'secret' | 'host'
}

export interface ComposeServiceData {
  id: string
  name: string
  type: 'ClusterIP' | 'LoadBalancer' | 'NodePort'
  clusterIP: string
  ports: ComposeServicePort[]
  volumes: ComposeServiceVolume[]
  dns: string
}

const topLevelKeys = new Set(['version', 'services', 'volumes', 'networks', 'configs', 'secrets'])

function stripQuotes(value: string): string {
  return value.trim().replace(/^['"]|['"]$/g, '')
}

function parsePortMapping(value: string): ComposeServicePort | null {
  const cleaned = stripQuotes(value)
  const singlePort = cleaned.match(/^(\d+)(?:\/(tcp|udp))?$/i)
  if (singlePort) {
    const port = Number.parseInt(singlePort[1], 10)
    return {
      port,
      targetPort: port,
      protocol: (singlePort[2] || 'TCP').toUpperCase(),
    }
  }

  const match = cleaned.match(/^(?:(?:[\d.]+|\[[^\]]+\]):)?(\d+)\s*:\s*(\d+)(?:\/(tcp|udp))?$/i)
  if (!match) return null

  return {
    port: Number.parseInt(match[1], 10),
    targetPort: Number.parseInt(match[2], 10),
    protocol: (match[3] || 'TCP').toUpperCase(),
  }
}

function parseVolumeMount(value: string): ComposeServiceVolume | null {
  const cleaned = stripQuotes(value)
  const parts = cleaned.split(':')
  if (parts.length < 2) return null

  const [name, mountPath] = parts
  if (!name || !mountPath) return null

  return {
    name,
    mountPath,
    type: name.startsWith('.') || name.startsWith('/') ? 'host' : 'persistent',
  }
}

export function parseComposeServices(composeContent: string, targetNamespace: string): ComposeServiceData[] {
  const parsed: ComposeServiceData[] = []
  const lines = composeContent.split('\n')
  let inServices = false
  let currentSvc: string | null = null
  let currentSection: 'ports' | 'volumes' | null = null
  let ports: ComposeServicePort[] = []
  let volumes: ComposeServiceVolume[] = []

  function pushCurrentService() {
    if (!currentSvc) return
    parsed.push({
      id: currentSvc,
      name: currentSvc,
      type: 'ClusterIP',
      clusterIP: '',
      dns: `${currentSvc}.${targetNamespace}.svc.cluster.local`,
      ports,
      volumes,
    })
  }

  for (const rawLine of lines) {
    const line = rawLine.replace(/\s+#.*$/, '')
    if (!line.trim()) continue

    const topLevel = line.match(/^([\w][\w-]*):\s*$/)
    if (topLevel) {
      if (inServices) pushCurrentService()
      inServices = topLevel[1] === 'services'
      currentSvc = null
      currentSection = null
      ports = []
      volumes = []
      continue
    }

    if (!inServices) continue

    const serviceMatch = line.match(/^\s{2}([\w][\w-]*):\s*$/)
    if (serviceMatch) {
      if (!topLevelKeys.has(serviceMatch[1])) {
        pushCurrentService()
        currentSvc = serviceMatch[1]
        currentSection = null
        ports = []
        volumes = []
      }
      continue
    }

    const sectionMatch = line.match(/^\s{4}(ports|volumes):\s*$/)
    if (sectionMatch) {
      currentSection = sectionMatch[1] as 'ports' | 'volumes'
      continue
    }

    const otherSectionMatch = line.match(/^\s{4}[\w][\w-]*:/)
    if (otherSectionMatch) {
      currentSection = null
      continue
    }

    const listMatch = line.match(/^\s{6}-\s*(.+?)\s*$/)
    if (!listMatch || !currentSvc) continue

    if (currentSection === 'ports') {
      const port = parsePortMapping(listMatch[1])
      if (port) ports.push(port)
      continue
    }

    if (currentSection === 'volumes') {
      const volume = parseVolumeMount(listMatch[1])
      if (volume) volumes.push(volume)
    }
  }

  if (inServices) pushCurrentService()
  return parsed
}
