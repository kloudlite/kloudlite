// Proxy server that handles WebSocket and forwards HTTP to Next.js standalone
const http = require('http')
const https = require('https')
const { WebSocket, WebSocketServer } = require('ws')

const port = parseInt(process.env.PORT || '3000', 10)
const nextPort = 3002 // Next.js runs on this port internally
const apiUrl = process.env.NEXT_PUBLIC_API_URL || process.env.API_URL || 'http://localhost:8080'

// Parse API URL for WebSocket proxy
const apiUrlParsed = new URL(apiUrl)
const wsTargetBase = apiUrlParsed.protocol === 'https:'
  ? `wss://${apiUrlParsed.host}`
  : `ws://${apiUrlParsed.host}`

// Helper functions for Kubernetes metrics
function parseCPU(cpuString) {
  if (cpuString.endsWith('n')) {
    return parseInt(cpuString.slice(0, -1)) / 1_000_000_000
  }
  if (cpuString.endsWith('u')) {
    return parseInt(cpuString.slice(0, -1)) / 1_000_000
  }
  if (cpuString.endsWith('m')) {
    return parseInt(cpuString.slice(0, -1))
  }
  return parseFloat(cpuString) * 1000
}

function parseMemory(memoryString) {
  const units = {
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

async function fetchNodeMetrics(nodeName) {
  try {
    const httpModule = apiUrlParsed.protocol === 'https:' ? https : http

    // Fetch node metrics
    const metricsData = await new Promise((resolve, reject) => {
      const req = httpModule.get(
        `${apiUrl}/apis/metrics.k8s.io/v1beta1/nodes/${nodeName}`,
        { rejectUnauthorized: false },
        (res) => {
          let data = ''
          res.on('data', chunk => data += chunk)
          res.on('end', () => {
            if (res.statusCode === 200) {
              resolve(JSON.parse(data))
            } else {
              reject(new Error(`Failed to fetch metrics: ${res.statusCode}`))
            }
          })
        }
      )
      req.on('error', reject)
    })

    // Fetch node details
    const nodeData = await new Promise((resolve, reject) => {
      const req = httpModule.get(
        `${apiUrl}/api/v1/nodes/${nodeName}`,
        { rejectUnauthorized: false },
        (res) => {
          let data = ''
          res.on('data', chunk => data += chunk)
          res.on('end', () => {
            if (res.statusCode === 200) {
              resolve(JSON.parse(data))
            } else {
              reject(new Error(`Failed to fetch node: ${res.statusCode}`))
            }
          })
        }
      )
      req.on('error', reject)
    })

    return {
      cpu: {
        usage: parseCPU(metricsData.usage.cpu),
        capacity: parseCPU(nodeData.status.capacity.cpu),
        allocatable: parseCPU(nodeData.status.allocatable.cpu),
      },
      memory: {
        usage: parseMemory(metricsData.usage.memory),
        capacity: parseMemory(nodeData.status.capacity.memory),
        allocatable: parseMemory(nodeData.status.allocatable.memory),
      },
      timestamp: metricsData.timestamp,
    }
  } catch (error) {
    console.error('Error fetching node metrics:', error.message)
    return null
  }
}

// WebSocket paths that should be proxied
const wsPathPatterns = [
  /^\/api\/v1\/environments\/[^/]+\/status-ws/,
  /^\/api\/v1\/namespaces\/[^/]+\/workspaces\/[^/]+\/status-ws/,
  /^\/api\/v1\/work-machines\/[^/]+\/metrics-ws/,
  /^\/api\/v1\/namespaces\/[^/]+\/services\/[^/]+\/logs-ws/,
  /^\/api\/v1\/resource-events-ws/,
]

function isWebSocketPath(pathname) {
  return wsPathPatterns.some(pattern => pattern.test(pathname))
}

// Create main server
const server = http.createServer((req, res) => {
  // Filter out hop-by-hop headers that shouldn't be forwarded
  const hopByHopHeaders = ['connection', 'keep-alive', 'transfer-encoding', 'te', 'trailer', 'upgrade']
  const forwardHeaders = {}
  for (const [key, value] of Object.entries(req.headers)) {
    if (!hopByHopHeaders.includes(key.toLowerCase())) {
      forwardHeaders[key] = value
    }
  }

  // Preserve original host in x-forwarded-host for Next.js Server Actions
  const originalHost = req.headers.host
  if (originalHost && !forwardHeaders['x-forwarded-host']) {
    forwardHeaders['x-forwarded-host'] = originalHost
  }

  // Set x-forwarded-proto if not already set
  if (!forwardHeaders['x-forwarded-proto']) {
    forwardHeaders['x-forwarded-proto'] = 'https'
  }

  // Set host header for internal routing to Next.js
  forwardHeaders['host'] = `127.0.0.1:${nextPort}`

  const options = {
    hostname: '127.0.0.1',
    port: nextPort,
    path: req.url,
    method: req.method,
    headers: forwardHeaders,
  }

  const proxyReq = http.request(options, (proxyRes) => {
    // Filter hop-by-hop headers from response too
    const responseHeaders = {}
    for (const [key, value] of Object.entries(proxyRes.headers)) {
      // Skip transfer-encoding as it will be set automatically when piping
      if (key.toLowerCase() !== 'transfer-encoding') {
        responseHeaders[key] = value
      }
    }
    res.writeHead(proxyRes.statusCode, responseHeaders)
    proxyRes.pipe(res)
  })

  proxyReq.on('error', (err) => {
    console.error('Proxy error:', err.message)
    if (!res.headersSent) {
      res.writeHead(502)
      res.end('Proxy error')
    }
  })

  req.pipe(proxyReq)
})

// WebSocket server for handling upgrades
const wss = new WebSocketServer({ noServer: true })

server.on('upgrade', (req, socket, head) => {
  const pathname = req.url.split('?')[0]

  if (isWebSocketPath(pathname)) {
    // Check if this is a metrics WebSocket request
    const metricsMatch = pathname.match(/^\/api\/v1\/work-machines\/([^/]+)\/metrics-ws/)

    // Resource events: bridge Next.js SSE endpoint → WebSocket
    const isResourceEvents = pathname === '/api/v1/resource-events-ws'

    if (isResourceEvents) {
      wss.handleUpgrade(req, socket, head, (clientWs) => {
        // Forward cookies to Next.js so getSession() works
        const headers = {}
        if (req.headers.cookie) headers['Cookie'] = req.headers.cookie

        const sseReq = http.get(
          `http://127.0.0.1:${nextPort}/api/resource-events`,
          { headers },
          (sseRes) => {
            if (sseRes.statusCode !== 200) {
              clientWs.close(1008, 'Upstream auth failed')
              sseRes.resume()
              return
            }

            let buffer = ''

            sseRes.on('data', (chunk) => {
              buffer += chunk.toString()
              // SSE format: "data: {...}\n\n"
              let idx
              while ((idx = buffer.indexOf('\n\n')) !== -1) {
                const frame = buffer.slice(0, idx)
                buffer = buffer.slice(idx + 2)
                for (const line of frame.split('\n')) {
                  if (line.startsWith('data: ')) {
                    const json = line.slice(6)
                    if (clientWs.readyState === WebSocket.OPEN) {
                      clientWs.send(json)
                    }
                  }
                }
              }
            })

            sseRes.on('end', () => {
              if (clientWs.readyState === WebSocket.OPEN) {
                clientWs.close(1001, 'Upstream closed')
              }
            })

            sseRes.on('error', (err) => {
              console.error('SSE bridge error:', err.message)
              if (clientWs.readyState === WebSocket.OPEN) {
                clientWs.close(1011, 'Upstream error')
              }
            })
          }
        )

        sseReq.on('error', (err) => {
          console.error('SSE bridge request error:', err.message)
          clientWs.close(1011, 'Connection failed')
        })

        clientWs.on('close', () => {
          sseReq.destroy()
        })

        clientWs.on('error', () => {
          sseReq.destroy()
        })
      })
    } else if (metricsMatch) {
      const nodeName = metricsMatch[1]
      console.log(`Metrics WebSocket: ${pathname} -> Kubernetes API for node ${nodeName}`)

      wss.handleUpgrade(req, socket, head, (clientWs) => {
        let interval = null

        const sendMetrics = async () => {
          const metrics = await fetchNodeMetrics(nodeName)
          if (metrics && clientWs.readyState === WebSocket.OPEN) {
            clientWs.send(JSON.stringify({
              type: 'metrics',
              nodeMetrics: metrics,
            }))
          }
        }

        // Send initial metrics
        sendMetrics()

        // Send metrics every 2 seconds
        interval = setInterval(sendMetrics, 2000)

        clientWs.on('close', () => {
          if (interval) clearInterval(interval)
        })

        clientWs.on('error', (err) => {
          console.error('Client WebSocket error:', err.message)
          if (interval) clearInterval(interval)
        })
      })
    } else {
      // Forward other WebSocket requests to backend
      wss.handleUpgrade(req, socket, head, (clientWs) => {
        const targetUrl = `${wsTargetBase}${req.url}`
        console.log(`WebSocket proxy: ${req.url} -> ${targetUrl}`)

        const backendWs = new WebSocket(targetUrl)

        backendWs.on('open', () => {
          console.log('Backend WebSocket connected')
        })

        backendWs.on('message', (data, isBinary) => {
          if (clientWs.readyState === WebSocket.OPEN) {
            clientWs.send(data, { binary: isBinary })
          }
        })

        backendWs.on('close', (code, reason) => {
          clientWs.close(code, reason)
        })

        backendWs.on('error', (err) => {
          console.error('Backend WebSocket error:', err.message)
          clientWs.close(1011, 'Backend error')
        })

        clientWs.on('message', (data, isBinary) => {
          if (backendWs.readyState === WebSocket.OPEN) {
            backendWs.send(data, { binary: isBinary })
          }
        })

        clientWs.on('close', () => {
          backendWs.close()
        })

        clientWs.on('error', (err) => {
          console.error('Client WebSocket error:', err.message)
          backendWs.close()
        })
      })
    }
  } else {
    // For non-API WebSocket (like HMR), just close
    socket.destroy()
  }
})

server.listen(port, '0.0.0.0', () => {
  console.log(`Proxy server listening on port ${port}`)
  console.log(`Forwarding HTTP to Next.js on port ${nextPort}`)
  console.log(`Proxying WebSocket to: ${wsTargetBase}`)
})
