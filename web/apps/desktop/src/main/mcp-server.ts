import { createServer, type IncomingMessage, type ServerResponse } from 'http'
import { BrowserWindow } from 'electron'
import { randomUUID } from 'crypto'

const MCP_PORT = 3789
const REQUEST_TIMEOUT = 30_000

interface PendingRequest {
  resolve: (result: unknown) => void
  reject: (err: Error) => void
  timer: ReturnType<typeof setTimeout>
}

const pending = new Map<string, PendingRequest>()

const toolDefinitions = [
  {
    name: 'browser_list_tabs',
    description: 'List all open browser tabs with their IDs, URLs, and titles',
    inputSchema: { type: 'object', properties: {} }
  },
  {
    name: 'browser_new_tab',
    description: 'Open a new browser tab at the given URL',
    inputSchema: {
      type: 'object',
      properties: {
        url: { type: 'string', description: 'URL to navigate to (defaults to google.com)' }
      }
    }
  },
  {
    name: 'browser_close_tab',
    description: 'Close a browser tab by ID',
    inputSchema: {
      type: 'object',
      properties: {
        tabId: { type: 'string', description: 'Tab ID to close' }
      },
      required: ['tabId']
    }
  },
  {
    name: 'browser_activate_tab',
    description: 'Switch to a specific tab by ID',
    inputSchema: {
      type: 'object',
      properties: {
        tabId: { type: 'string', description: 'Tab ID to switch to' }
      },
      required: ['tabId']
    }
  },
  {
    name: 'browser_navigate',
    description: 'Navigate the active or specified tab to a URL',
    inputSchema: {
      type: 'object',
      properties: {
        url: { type: 'string', description: 'URL to navigate to' },
        tabId: { type: 'string', description: 'Optional tab ID (uses active tab if omitted)' }
      },
      required: ['url']
    }
  },
  {
    name: 'browser_go_back',
    description: 'Go back in tab history',
    inputSchema: {
      type: 'object',
      properties: {
        tabId: { type: 'string', description: 'Optional tab ID (uses active tab if omitted)' }
      }
    }
  },
  {
    name: 'browser_go_forward',
    description: 'Go forward in tab history',
    inputSchema: {
      type: 'object',
      properties: {
        tabId: { type: 'string', description: 'Optional tab ID (uses active tab if omitted)' }
      }
    }
  },
  {
    name: 'browser_reload',
    description: 'Reload the current page in the specified tab',
    inputSchema: {
      type: 'object',
      properties: {
        tabId: { type: 'string', description: 'Optional tab ID (uses active tab if omitted)' }
      }
    }
  },
  {
    name: 'browser_screenshot',
    description: 'Take a screenshot of the current page (returns base64 PNG)',
    inputSchema: {
      type: 'object',
      properties: {
        tabId: { type: 'string', description: 'Optional tab ID (uses active tab if omitted)' }
      }
    }
  },
  {
    name: 'browser_click',
    description: 'Click an element on the page by CSS selector',
    inputSchema: {
      type: 'object',
      properties: {
        selector: { type: 'string', description: 'CSS selector of the element to click' },
        tabId: { type: 'string', description: 'Optional tab ID (uses active tab if omitted)' }
      },
      required: ['selector']
    }
  },
  {
    name: 'browser_fill',
    description: 'Fill a form field by CSS selector',
    inputSchema: {
      type: 'object',
      properties: {
        selector: { type: 'string', description: 'CSS selector of the input element' },
        value: { type: 'string', description: 'Value to fill' },
        tabId: { type: 'string', description: 'Optional tab ID (uses active tab if omitted)' }
      },
      required: ['selector', 'value']
    }
  },
  {
    name: 'browser_get_text',
    description: 'Get visible text from the page, optionally scoped to a selector',
    inputSchema: {
      type: 'object',
      properties: {
        selector: { type: 'string', description: 'Optional CSS selector to scope text extraction' },
        tabId: { type: 'string', description: 'Optional tab ID (uses active tab if omitted)' }
      }
    }
  },
  {
    name: 'browser_evaluate',
    description: 'Execute arbitrary JavaScript in the page context',
    inputSchema: {
      type: 'object',
      properties: {
        script: { type: 'string', description: 'JavaScript code to execute' },
        tabId: { type: 'string', description: 'Optional tab ID (uses active tab if omitted)' }
      },
      required: ['script']
    }
  },
  {
    name: 'browser_get_url',
    description: 'Get the current URL of a tab',
    inputSchema: {
      type: 'object',
      properties: {
        tabId: { type: 'string', description: 'Optional tab ID (uses active tab if omitted)' }
      }
    }
  },
  {
    name: 'browser_get_title',
    description: 'Get the page title of a tab',
    inputSchema: {
      type: 'object',
      properties: {
        tabId: { type: 'string', description: 'Optional tab ID (uses active tab if omitted)' }
      }
    }
  }
]

function jsonRPCError(id: unknown, code: number, message: string) {
  return { jsonrpc: '2.0', id, error: { code, message } }
}

function jsonRPCSuccess(id: unknown, result: unknown) {
  return { jsonrpc: '2.0', id, result }
}

function sendToRenderer(command: string, args: Record<string, unknown>): Promise<unknown> {
  const win = BrowserWindow.getAllWindows()[0]
  if (!win) throw new Error('No browser window available')

  const requestId = randomUUID()
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      if (pending.has(requestId)) {
        pending.delete(requestId)
        reject(new Error('Command timed out'))
      }
    }, REQUEST_TIMEOUT)

    pending.set(requestId, { resolve, reject, timer })
    win.webContents.send('mcp-command', { requestId, command, args })
  })
}

function handleToolCall(name: string, args: Record<string, unknown>): Promise<unknown> {
  // Commands that don't need renderer — handled directly
  if (name === 'browser_list_tabs') {
    return sendToRenderer('list_tabs', {})
  }
  if (name === 'browser_new_tab') {
    return sendToRenderer('new_tab', { url: (args.url as string) || 'https://google.com' })
  }
  if (name === 'browser_close_tab') {
    return sendToRenderer('close_tab', { tabId: args.tabId as string })
  }
  if (name === 'browser_activate_tab') {
    return sendToRenderer('activate_tab', { tabId: args.tabId as string })
  }

  // Commands that need renderer
  return sendToRenderer(name, args as Record<string, string>)
}

function handleJSONRPC(body: Record<string, unknown>, res: ServerResponse) {
  const { id, method, params } = body

  if (method === 'tools/list') {
    res.end(JSON.stringify(jsonRPCSuccess(id, { tools: toolDefinitions })))
    return
  }

  if (method === 'tools/call') {
    const p = params as { name?: string; arguments?: Record<string, unknown> } | undefined
    const toolName = p?.name
    const toolArgs = p?.arguments || {}

    const tool = toolDefinitions.find(t => t.name === toolName)
    if (!tool) {
      res.end(JSON.stringify(jsonRPCError(id, -32602, `Unknown tool: ${toolName}`)))
      return
    }

    handleToolCall(tool.name, toolArgs)
      .then((result) => {
        const content: unknown[] = []

        if (typeof result === 'string') {
          content.push({ type: 'text', text: result })
        } else if (result && typeof result === 'object' && 'type' in (result as any)) {
          content.push(result)
        } else {
          content.push({ type: 'text', text: JSON.stringify(result, null, 2) })
        }

        res.end(JSON.stringify(jsonRPCSuccess(id, { content })))
      })
      .catch((err) => {
        res.end(JSON.stringify(jsonRPCError(id, -32603, err.message)))
      })
    return
  }

  if (method === 'tools/description') {
    res.end(JSON.stringify(jsonRPCSuccess(id, { tools: toolDefinitions })))
    return
  }

  res.end(JSON.stringify(jsonRPCError(id, -32601, `Method not found: ${method}`)))
}

function parseBody(req: IncomingMessage): Promise<Record<string, unknown>> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = []
    req.on('data', (c: Buffer) => chunks.push(c))
    req.on('end', () => {
      try {
        resolve(JSON.parse(Buffer.concat(chunks).toString()))
      } catch (e) {
        reject(e)
      }
    })
    req.on('error', reject)
  })
}

let server: ReturnType<typeof createServer> | null = null

export function startMCPServer() {
  server = createServer(async (req, res) => {
    res.setHeader('Access-Control-Allow-Origin', '*')
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type')

    if (req.method === 'OPTIONS') {
      res.writeHead(204)
      res.end()
      return
    }

    // SSE endpoint for MCP
    if (req.method === 'GET' && req.url === '/sse') {
      const sessionId = randomUUID()
      res.writeHead(200, {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache',
        Connection: 'keep-alive',
        'Access-Control-Allow-Origin': '*'
      })

      res.write(`event: endpoint\ndata: /messages?sessionId=${sessionId}\n\n`)

      // Keep alive
      const keepAlive = setInterval(() => {
        res.write(': keepalive\n\n')
      }, 15000)

      req.on('close', () => {
        clearInterval(keepAlive)
      })
      return
    }

    // POST /messages — receive JSON-RPC from client
    if (req.method === 'POST' && req.url?.startsWith('/messages')) {
      try {
        const body = await parseBody(req)
        res.writeHead(200, { 'Content-Type': 'application/json' })
        handleJSONRPC(body, res)
      } catch {
        res.writeHead(400, { 'Content-Type': 'application/json' })
        res.end(JSON.stringify(jsonRPCError(null, -32700, 'Parse error')))
      }
      return
    }

    // POST /mcp — simpler endpoint for HTTP-based MCP clients
    if (req.method === 'POST' && req.url === '/mcp') {
      try {
        const body = await parseBody(req)
        res.writeHead(200, { 'Content-Type': 'application/json' })
        handleJSONRPC(body, res)
      } catch {
        res.writeHead(400, { 'Content-Type': 'application/json' })
        res.end(JSON.stringify(jsonRPCError(null, -32700, 'Parse error')))
      }
      return
    }

    // GET / — health check / info
    if (req.method === 'GET' && req.url === '/') {
      res.writeHead(200, { 'Content-Type': 'application/json' })
      res.end(JSON.stringify({
        server: 'kloudlite-browser-mcp',
        version: '0.1.0',
        tools: toolDefinitions.length,
        port: MCP_PORT
      }))
      return
    }

    res.writeHead(404)
    res.end('Not found')
  })

  server.listen(MCP_PORT, () => {
    console.log(`[mcp-server] browser MCP server listening on port ${MCP_PORT}`)
  })
}

export function stopMCPServer() {
  if (server) {
    server.close()
    server = null
  }
  for (const [, p] of pending) {
    clearTimeout(p.timer)
    p.reject(new Error('Server shutting down'))
  }
  pending.clear()
}

// Receive result from renderer
export function receiveMCPResult(requestId: string, result: unknown, error: string | null) {
  const p = pending.get(requestId)
  if (!p) return

  clearTimeout(p.timer)
  pending.delete(requestId)

  if (error) {
    p.reject(new Error(error))
  } else {
    p.resolve(result)
  }
}
