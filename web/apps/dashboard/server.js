const { createServer } = require('http')
const { parse } = require('url')
const next = require('next')
const { WebSocket, WebSocketServer } = require('ws')

const dev = process.env.NODE_ENV !== 'production'
const hostname = process.env.HOSTNAME || '0.0.0.0'
const port = parseInt(process.env.PORT || '3000', 10)
const apiUrl = process.env.API_URL || 'http://api-server.kloudlite.svc.cluster.local'

// Parse API URL for WebSocket proxy
const apiUrlParsed = new URL(apiUrl)
const wsTargetBase = apiUrlParsed.protocol === 'https:'
  ? `wss://${apiUrlParsed.host}`
  : `ws://${apiUrlParsed.host}`

// Create Next.js app
const app = next({ dev, hostname, port })
const handle = app.getRequestHandler()

// WebSocket paths that should be proxied to the API server
const wsPathPatterns = [
  /^\/api\/v1\/environments\/[^/]+\/status-ws/,
  /^\/api\/v1\/namespaces\/[^/]+\/workspaces\/[^/]+\/status-ws/,
  /^\/api\/v1\/work-machines\/[^/]+\/metrics-ws/,
  /^\/api\/v1\/namespaces\/[^/]+\/services\/[^/]+\/logs-ws/,
]

function isWebSocketPath(pathname) {
  return wsPathPatterns.some(pattern => pattern.test(pathname))
}

app.prepare().then(() => {
  const server = createServer((req, res) => {
    const parsedUrl = parse(req.url, true)
    handle(req, res, parsedUrl)
  })

  // Create WebSocket server for handling upgrades
  const wss = new WebSocketServer({ noServer: true })

  // Handle WebSocket upgrade requests
  server.on('upgrade', (req, socket, head) => {
    const parsedUrl = parse(req.url, true)

    if (isWebSocketPath(parsedUrl.pathname)) {
      wss.handleUpgrade(req, socket, head, (clientWs) => {
        // Build target URL with query string
        const targetUrl = `${wsTargetBase}${req.url}`
        console.log(`WebSocket proxy: ${req.url} -> ${targetUrl}`)

        // Connect to backend
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
      socket.destroy()
    }
  })

  server.listen(port, hostname, () => {
    console.log(`> Ready on http://${hostname}:${port}`)
    console.log(`> WebSocket proxy target: ${wsTargetBase}`)
  })
})
