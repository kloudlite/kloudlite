'use client'

import { Pencil, Trash2, Plus, ChevronRight, FileText, Settings } from 'lucide-react'

// Preview frame wrapper with browser chrome
function PreviewFrame({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative rounded-lg overflow-hidden shadow-lg border border-border/50">
      {/* Browser chrome */}
      <div className="bg-zinc-800 px-4 py-2.5 flex items-center gap-3">
        <div className="flex gap-1.5">
          <div className="w-3 h-3 rounded-full bg-[#ff5f57]" />
          <div className="w-3 h-3 rounded-full bg-[#febc2e]" />
          <div className="w-3 h-3 rounded-full bg-[#28c840]" />
        </div>
        <div className="flex-1 flex justify-center">
          <div className="bg-zinc-700/50 rounded px-3 py-1 text-zinc-400 text-[10px] flex items-center gap-2">
            <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            console.kloudlite.io
          </div>
        </div>
        <div className="text-[9px] font-medium text-zinc-500 bg-zinc-700/50 px-2 py-0.5 rounded">
          PREVIEW
        </div>
      </div>
      {/* Content */}
      <div className="bg-card">
        {children}
      </div>
    </div>
  )
}

const envVars = [
  { key: 'API_URL', value: 'https://api.example.com', type: 'config' as const },
  { key: 'LOG_LEVEL', value: 'debug', type: 'config' as const },
  { key: 'DB_PASSWORD', value: '••••••••', type: 'secret' as const },
  { key: 'JWT_SECRET', value: '••••••••', type: 'secret' as const },
]

export function EnvVarsPreview() {
  return (
    <PreviewFrame>
      <div className="text-xs">
        {/* Dashboard Header */}
        <div className="bg-background border-b px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-1.5 text-muted-foreground">
            <span>Environments</span>
            <ChevronRight className="h-3 w-3" />
            <span>staging</span>
            <ChevronRight className="h-3 w-3" />
            <span className="text-foreground font-medium">Configs & Secrets</span>
          </div>
          <span className="font-bold text-sm">Kloudlite</span>
        </div>

        {/* Tabs */}
        <div className="bg-background border-b px-4 flex gap-4">
          <button className="py-2.5 border-b-2 border-primary text-primary font-medium flex items-center gap-1.5">
            <Settings className="h-3.5 w-3.5" />
            Environment Variables
          </button>
          <button className="py-2.5 border-b-2 border-transparent text-muted-foreground flex items-center gap-1.5">
            <FileText className="h-3.5 w-3.5" />
            Config Files
          </button>
        </div>

        {/* Content */}
        <div className="p-4">
          <div className="flex items-center justify-between mb-4">
            <p className="text-muted-foreground text-[11px]">
              Manage environment variables for your services
            </p>
            <button className="bg-primary text-primary-foreground px-3 py-1.5 text-[11px] font-medium flex items-center gap-1 rounded">
              <Plus className="h-3 w-3" />
              Add Variable
            </button>
          </div>

          <div className="border rounded overflow-hidden">
            <table className="w-full">
              <thead>
                <tr className="bg-muted/30">
                  <th className="text-left px-3 py-2 font-medium text-muted-foreground">Key</th>
                  <th className="text-left px-3 py-2 font-medium text-muted-foreground">Value</th>
                  <th className="text-left px-3 py-2 font-medium text-muted-foreground">Type</th>
                  <th className="text-right px-3 py-2 font-medium text-muted-foreground"></th>
                </tr>
              </thead>
              <tbody>
                {envVars.map((item) => (
                  <tr key={item.key} className="border-t">
                    <td className="px-3 py-2 font-mono">{item.key}</td>
                    <td className="px-3 py-2 font-mono text-muted-foreground">{item.value}</td>
                    <td className="px-3 py-2">
                      <span
                        className={`px-2 py-0.5 text-[10px] font-medium rounded ${
                          item.type === 'config'
                            ? 'bg-primary/10 text-primary'
                            : 'bg-purple-500/10 text-purple-500'
                        }`}
                      >
                        {item.type === 'config' ? 'Config' : 'Secret'}
                      </span>
                    </td>
                    <td className="px-3 py-2">
                      <div className="flex items-center justify-end gap-1">
                        <button className="p-1 hover:bg-muted rounded">
                          <Pencil className="h-3 w-3 text-muted-foreground" />
                        </button>
                        <button className="p-1 hover:bg-muted rounded">
                          <Trash2 className="h-3 w-3 text-muted-foreground" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </PreviewFrame>
  )
}

const configFiles = [
  { name: 'nginx.conf' },
  { name: 'app-settings.json' },
]

export function ConfigFilesPreview() {
  return (
    <PreviewFrame>
      <div className="text-xs">
        {/* Dashboard Header */}
        <div className="bg-background border-b px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-1.5 text-muted-foreground">
            <span>Environments</span>
            <ChevronRight className="h-3 w-3" />
            <span>staging</span>
            <ChevronRight className="h-3 w-3" />
            <span className="text-foreground font-medium">Configs & Secrets</span>
          </div>
          <span className="font-bold text-sm">Kloudlite</span>
        </div>

        {/* Tabs */}
        <div className="bg-background border-b px-4 flex gap-4">
          <button className="py-2.5 border-b-2 border-transparent text-muted-foreground flex items-center gap-1.5">
            <Settings className="h-3.5 w-3.5" />
            Environment Variables
          </button>
          <button className="py-2.5 border-b-2 border-primary text-primary font-medium flex items-center gap-1.5">
            <FileText className="h-3.5 w-3.5" />
            Config Files
          </button>
        </div>

        {/* Content */}
        <div className="p-4">
          <div className="flex items-center justify-between mb-4">
            <p className="text-muted-foreground text-[11px]">
              Upload configuration files to mount into containers
            </p>
            <button className="bg-primary text-primary-foreground px-3 py-1.5 text-[11px] font-medium flex items-center gap-1 rounded">
              <Plus className="h-3 w-3" />
              Upload File
            </button>
          </div>

          <div className="border rounded overflow-hidden">
            <table className="w-full">
              <thead>
                <tr className="bg-muted/30">
                  <th className="text-left px-3 py-2 font-medium text-muted-foreground">File Name</th>
                  <th className="text-right px-3 py-2 font-medium text-muted-foreground"></th>
                </tr>
              </thead>
              <tbody>
                {configFiles.map((file) => (
                  <tr key={file.name} className="border-t">
                    <td className="px-3 py-2 font-mono flex items-center gap-2">
                      <FileText className="h-3.5 w-3.5 text-muted-foreground" />
                      {file.name}
                    </td>
                    <td className="px-3 py-2">
                      <div className="flex items-center justify-end gap-1">
                        <button className="p-1 hover:bg-muted rounded">
                          <Pencil className="h-3 w-3 text-muted-foreground" />
                        </button>
                        <button className="p-1 hover:bg-muted rounded">
                          <Trash2 className="h-3 w-3 text-muted-foreground" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </PreviewFrame>
  )
}
