import { StatusBadge, StatusIndicator, StatusBar } from "@/components/molecules";
import { Activity } from "lucide-react";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function StatusPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Status Components
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Components for displaying system and resource status.
        </p>
      </div>

      <ComponentShowcase
        title="Status Badges"
        description="Badges with status indicators"
      >
        <div className="space-y-4">
          <div>
            <p className="text-sm font-medium mb-2">Status Types</p>
            <div className="flex flex-wrap gap-2">
              <StatusBadge status="running" />
              <StatusBadge status="pending" />
              <StatusBadge status="stopped" />
              <StatusBadge status="warning" />
              <StatusBadge status="error" />
              <StatusBadge status="success" />
            </div>
          </div>

          <div>
            <p className="text-sm font-medium mb-2">Sizes</p>
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge status="running" size="sm" />
              <StatusBadge status="running" size="default" />
              <StatusBadge status="running" size="lg" />
            </div>
          </div>

          <div>
            <p className="text-sm font-medium mb-2">With Pulse Animation</p>
            <div className="flex flex-wrap gap-2">
              <StatusBadge status="running" showPulse />
              <StatusBadge status="pending" showPulse />
              <StatusBadge status="warning" showPulse />
            </div>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Status Indicators"
        description="Simple status indicator dots"
      >
        <div className="space-y-4">
          <div>
            <p className="text-sm font-medium mb-2">Indicator Types</p>
            <div className="flex flex-wrap items-center gap-4">
              <div className="flex items-center gap-2">
                <StatusIndicator status="running" />
                <span className="text-sm">Running</span>
              </div>
              <div className="flex items-center gap-2">
                <StatusIndicator status="pending" />
                <span className="text-sm">Pending</span>
              </div>
              <div className="flex items-center gap-2">
                <StatusIndicator status="stopped" />
                <span className="text-sm">Stopped</span>
              </div>
              <div className="flex items-center gap-2">
                <StatusIndicator status="warning" />
                <span className="text-sm">Warning</span>
              </div>
              <div className="flex items-center gap-2">
                <StatusIndicator status="error" />
                <span className="text-sm">Error</span>
              </div>
            </div>
          </div>

          <div>
            <p className="text-sm font-medium mb-2">Sizes</p>
            <div className="flex flex-wrap items-center gap-4">
              <StatusIndicator status="running" size="sm" />
              <StatusIndicator status="running" size="default" />
              <StatusIndicator status="running" size="lg" />
            </div>
          </div>

          <div>
            <p className="text-sm font-medium mb-2">With Pulse</p>
            <div className="flex flex-wrap items-center gap-4">
              <StatusIndicator status="running" showPulse />
              <StatusIndicator status="warning" showPulse />
              <StatusIndicator status="error" showPulse />
            </div>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Status Bar"
        description="Full-width status bar component"
        padding="none"
      >
        <div className="space-y-4">
          <StatusBar
            statusText="All Systems Operational"
            statusType="success"
            metrics={[
              { label: "active environments", value: "42" },
              { label: "running workspaces", value: "128" },
              { label: "uptime", value: "99.9%" }
            ]}
            actionText="View detailed status"
            actionHref="#"
          />

          <StatusBar
            statusText="Maintenance in Progress"
            statusType="warning"
            statusIcon={Activity}
            metrics={[
              { label: "affected services", value: "3" },
              { label: "estimated time", value: "30 min" }
            ]}
            actionText="View maintenance details"
            actionHref="#"
          />

          <StatusBar
            statusText="System Degraded"
            statusType="error"
            metrics={[
              { label: "failed services", value: "2" },
              { label: "response time", value: "2.5s" }
            ]}
            actionText="View incidents"
            actionHref="#"
          />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Combined Examples"
        description="Status components used in context"
      >
        <div className="space-y-4">
          <div className="flex items-center justify-between p-4 border rounded-lg">
            <div className="flex items-center gap-3">
              <StatusIndicator status="running" showPulse />
              <div>
                <p className="font-medium">Production Environment</p>
                <p className="text-sm text-slate-600 dark:text-slate-400">
                  All services operational
                </p>
              </div>
            </div>
            <StatusBadge status="running" />
          </div>

          <div className="flex items-center justify-between p-4 border rounded-lg">
            <div className="flex items-center gap-3">
              <StatusIndicator status="warning" showPulse />
              <div>
                <p className="font-medium">Staging Environment</p>
                <p className="text-sm text-slate-600 dark:text-slate-400">
                  Database migration in progress
                </p>
              </div>
            </div>
            <StatusBadge status="warning" />
          </div>

          <div className="flex items-center justify-between p-4 border rounded-lg">
            <div className="flex items-center gap-3">
              <StatusIndicator status="stopped" />
              <div>
                <p className="font-medium">Development Environment</p>
                <p className="text-sm text-slate-600 dark:text-slate-400">
                  Environment paused to save resources
                </p>
              </div>
            </div>
            <StatusBadge status="stopped" />
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}