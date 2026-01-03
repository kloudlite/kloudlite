// Proxy server that handles WebSocket and forwards HTTP to Next.js standalone
const http = require('http')
const { WebSocket, WebSocketServer } = require('ws')

const port = parseInt(process.env.PORT || '3000', 10)
const nextPort = 3002 // Next.js runs on this port internally
const apiUrl = process.env.API_URL || 'http://api-server.kloudlite.svc.cluster.local'

// Parse API URL for WebSocket proxy
const apiUrlParsed = new URL(apiUrl)
const wsTargetBase = apiUrlParsed.protocol === 'https:'
  ? `wss://${apiUrlParsed.host}`
  : `ws://${apiUrlParsed.host}`

// WebSocket paths that should be proxied
const wsPathPatterns = [
  /^\/api\/v1\/environments\/[^/]+\/status-ws/,
  /^\/api\/v1\/namespaces\/[^/]+\/workspaces\/[^/]+\/status-ws/,
  /^\/api\/v1\/work-machines\/[^/]+\/metrics-ws/,
  /^\/api\/v1\/namespaces\/[^/]+\/services\/[^/]+\/logs-ws/,
]

function isWebSocketPath(pathname) {
  return wsPathPatterns.some(pattern => pattern.test(pathname))
}

// Create main server
const server = http.createServer((req, res) => {
  // Forward all HTTP requests to Next.js
  const options = {
    hostname: '127.0.0.1',
    port: nextPort,
    path: req.url,
    method: req.method,
    headers: req.headers,
  }

  const proxyReq = http.request(options, (proxyRes) => {
    res.writeHead(proxyRes.statusCode, proxyRes.headers)
    proxyRes.pipe(res)
  })

  proxyReq.on('error', (err) => {
    console.error('Proxy error:', err.message)
    res.writeHead(502)
    res.end('Proxy error')
  })

  req.pipe(proxyReq)
})

// WebSocket server for handling upgrades
const wss = new WebSocketServer({ noServer: true })

server.on('upgrade', (req, socket, head) => {
  const pathname = req.url.split('?')[0]

  if (isWebSocketPath(pathname)) {
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
