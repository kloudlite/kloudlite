import { Server, Boxes, Database, Code2 } from 'lucide-react'

export function ArchitectureDiagram() {
  return (
    <section className="mb-12 sm:mb-16">
      <div className="bg-gradient-to-br from-slate-50 to-blue-50/30 dark:from-slate-900/50 dark:to-slate-800/50 rounded-2xl border border-slate-200 dark:border-slate-700 p-8 sm:p-12">
        <div className="max-w-5xl mx-auto">

          {/* Control Node - Top Level */}
          <div className="flex flex-col items-center mb-8">
            <div className="relative">
              <div className="absolute -inset-1 bg-gradient-to-r from-blue-600 to-cyan-600 rounded-2xl blur opacity-25"></div>
              <div className="relative bg-white dark:bg-slate-950 rounded-2xl border-2 border-blue-500 shadow-xl px-12 py-8">
                <div className="absolute -top-4 left-6 bg-blue-600 text-white rounded-lg px-4 py-1.5 text-xs font-bold uppercase tracking-wide shadow-lg">
                  Control Plane
                </div>
                <div className="flex items-center gap-4">
                  <div className="bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl p-4 shadow-lg">
                    <Server className="h-10 w-10 text-white" />
                  </div>
                  <div>
                    <h4 className="text-slate-900 dark:text-slate-100 text-2xl font-bold mb-1">
                      Control Node
                    </h4>
                    <p className="text-slate-500 dark:text-slate-400 text-sm font-mono">
                      {'{subdomain}'}.khost.dev
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Connection Flow - Vertical with branching */}
          <div className="flex flex-col items-center mb-8">
            <div className="relative flex flex-col items-center">
              {/* Vertical line */}
              <div className="w-0.5 h-12 bg-gradient-to-b from-blue-500 to-purple-500 opacity-60"></div>
              {/* Horizontal branching */}
              <div className="relative w-64 h-0.5 bg-gradient-to-r from-purple-500 via-purple-500 to-purple-500 opacity-60"></div>
              {/* Vertical drops */}
              <div className="absolute top-12 left-1/2 -translate-x-1/2 flex justify-between w-full max-w-3xl">
                <div className="w-0.5 h-8 bg-purple-600 opacity-60"></div>
                <div className="w-0.5 h-8 bg-purple-600 opacity-60"></div>
                <div className="w-0.5 h-8 bg-purple-600 opacity-60"></div>
              </div>
            </div>
          </div>

          {/* Multiple Workmachines with interconnections */}
          <div className="relative mt-12">
            {/* Horizontal interconnection lines between workmachines */}
            <div className="hidden sm:block absolute top-1/2 left-0 right-0 -translate-y-1/2 h-0.5 bg-gradient-to-r from-purple-400 via-purple-500 to-purple-400 opacity-40"></div>

            {/* Connection dots on each workmachine */}
            <div className="hidden sm:flex absolute top-1/2 left-0 right-0 -translate-y-1/2 justify-between px-[16.66%]">
              <div className="w-2 h-2 rounded-full bg-purple-500 ring-2 ring-purple-300 dark:ring-purple-700"></div>
              <div className="w-2 h-2 rounded-full bg-purple-500 ring-2 ring-purple-300 dark:ring-purple-700"></div>
              <div className="w-2 h-2 rounded-full bg-purple-500 ring-2 ring-purple-300 dark:ring-purple-700"></div>
            </div>

            <div className="space-y-4 sm:space-y-0 sm:grid sm:grid-cols-3 gap-4 relative z-10">
              {/* Workmachine 1 */}
              <WorkmachineCard number={1} />

              {/* Workmachine 2 */}
              <WorkmachineCard number={2} />

              {/* Workmachine 3 */}
              <WorkmachineCard number={3} />
            </div>
          </div>

        </div>

        {/* Flow Legend - Outside the main container */}
        <div className="mt-8">
          <div className="flex flex-wrap gap-6 justify-center items-center text-sm">
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-full bg-gradient-to-r from-blue-500 to-purple-500"></div>
              <span className="text-slate-600 dark:text-slate-400 font-medium">Orchestration Flow</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-0.5 w-8 bg-purple-400"></div>
              <span className="text-slate-600 dark:text-slate-400 font-medium">Workmachine Network</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-full bg-purple-500"></div>
              <span className="text-slate-600 dark:text-slate-400 font-medium">Workmachine</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-full bg-emerald-500"></div>
              <span className="text-slate-600 dark:text-slate-400 font-medium">Workspaces</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-full bg-amber-500"></div>
              <span className="text-slate-600 dark:text-slate-400 font-medium">Environments</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

// Reusable Workmachine Card Component
function WorkmachineCard({ number }: { number: number }) {
  return (
    <div className="relative">
      <div className="absolute -inset-0.5 bg-gradient-to-r from-purple-600 to-pink-600 rounded-xl blur opacity-20"></div>
      <div className="relative bg-white dark:bg-slate-950 rounded-xl border-2 border-purple-500 p-4 shadow-lg">
        <div className="absolute -top-3 left-3 bg-purple-600 text-white rounded-md px-3 py-0.5 text-xs font-bold uppercase tracking-wide shadow-md">
          VM {number}
        </div>

        <div className="flex items-center gap-2 mb-4 mt-2">
          <div className="bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg p-2 shadow-md">
            <Boxes className="h-5 w-5 text-white" />
          </div>
          <h5 className="text-slate-900 dark:text-slate-100 text-sm font-bold">
            Workmachine
          </h5>
        </div>

        {/* Workspace */}
        <div className="mb-3 p-3 bg-gradient-to-br from-emerald-50 to-green-50 dark:from-emerald-950/50 dark:to-green-950/50 rounded-lg border border-emerald-300 dark:border-emerald-700">
          <div className="flex items-center gap-2 mb-1">
            <Code2 className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
            <h6 className="text-emerald-800 dark:text-emerald-300 text-xs font-semibold">Workspaces</h6>
          </div>
          <p className="text-emerald-600 dark:text-emerald-500 text-xs">Dev Containers</p>
        </div>

        {/* Environment */}
        <div className="p-3 bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-950/50 dark:to-orange-950/50 rounded-lg border border-amber-300 dark:border-amber-700">
          <div className="flex items-center gap-2 mb-1">
            <Database className="h-4 w-4 text-amber-600 dark:text-amber-400" />
            <h6 className="text-amber-800 dark:text-amber-300 text-xs font-semibold">Environments</h6>
          </div>
          <p className="text-amber-600 dark:text-amber-500 text-xs">Services & Apps</p>
        </div>
      </div>
    </div>
  )
}
