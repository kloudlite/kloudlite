"use client";

import { useState } from "react";
import { Button, Badge, IconBox } from "@/components/atoms";
import { BaseCard, EnvironmentHeader } from "@/components/organisms";
import { 
  Search, 
  Bell, 
  Settings, 
  User,
  ChevronDown,
  Plus,
  Cloud,
  Activity,
  Users,
  Shield,
  Terminal,
  GitBranch,
  Clock,
  AlertCircle,
  CheckCircle2,
  XCircle,
  Pause,
  Play,
  RefreshCw,
  MoreHorizontal,
  Download,
  Upload,
  Share2,
  Filter,
  SortDesc
} from "lucide-react";
import { ComponentShowcase } from "../../_components/component-showcase";
import { StatusBadge, StatusIndicator } from "@/components/molecules";

export default function HeadersPage() {
  const [searchQuery, setSearchQuery] = useState("");

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Headers
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Header components for different page contexts.
        </p>
      </div>

      <ComponentShowcase
        title="Basic Page Header"
        description="Simple header with title and actions"
        
      >
        <div className="bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800 p-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold text-slate-900 dark:text-white">Dashboard</h1>
              <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
                Welcome back! Here's what's happening with your projects.
              </p>
            </div>
            <div className="flex items-center gap-3">
              <Button variant="outline" size="sm">
                <Download className="h-4 w-4 mr-2" />
                Export
              </Button>
              <Button size="sm">
                <Plus className="h-4 w-4 mr-2" />
                New Project
              </Button>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Header with Search"
        description="Header with integrated search functionality"
        
      >
        <div className="bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800 p-6">
          <div className="flex items-center justify-between gap-6">
            <div className="flex-1 max-w-md">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search environments..."
                  className="w-full pl-10 pr-4 py-2 bg-slate-100 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Button variant="ghost" size="icon">
                <Bell className="h-5 w-5" />
              </Button>
              <Button variant="ghost" size="icon">
                <Settings className="h-5 w-5" />
              </Button>
              <div className="h-8 w-8 rounded-full bg-gradient-to-br from-blue-500 to-cyan-600 flex items-center justify-center text-white text-sm font-medium">
                JD
              </div>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Header with Tabs"
        description="Header combined with tab navigation"
        
      >
        <div className="bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800">
          <div className="px-6 pt-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h1 className="text-2xl font-semibold text-slate-900 dark:text-white">Team Settings</h1>
                <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
                  Manage your team configuration and preferences
                </p>
              </div>
              <Button>Save Changes</Button>
            </div>
          </div>
          <div className="px-6">
            <nav className="flex gap-6 border-b border-slate-200 dark:border-slate-800">
              <button className="pb-3 px-1 border-b-2 border-blue-600 text-sm font-medium text-blue-600">
                General
              </button>
              <button className="pb-3 px-1 border-b-2 border-transparent text-sm font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white">
                Members
              </button>
              <button className="pb-3 px-1 border-b-2 border-transparent text-sm font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white">
                Security
              </button>
              <button className="pb-3 px-1 border-b-2 border-transparent text-sm font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white">
                Billing
              </button>
            </nav>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Environment Header"
        description="Specialized header for environment details"
        
      >
        <EnvironmentHeader
          environment={{
            name: "production",
            status: "running",
            region: "us-west-2",
            lastUpdated: "2 hours ago"
          }}
          onAction={(action) => console.log('Action:', action)}
        />
      </ComponentShowcase>

      <ComponentShowcase
        title="Header with Stats"
        description="Header displaying key metrics"
        
      >
        <div className="bg-gradient-to-r from-slate-900 to-slate-800 dark:from-slate-950 dark:to-slate-900 p-6">
          <div className="mb-6">
            <h1 className="text-2xl font-semibold text-white">Infrastructure Overview</h1>
            <p className="text-sm text-slate-300 mt-1">
              Real-time status of your cloud resources
            </p>
          </div>
          <div className="grid grid-cols-4 gap-4">
            <div className="bg-white/10 backdrop-blur rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <Cloud className="h-5 w-5 text-blue-400" />
                <StatusIndicator status="running" showPulse />
              </div>
              <p className="text-2xl font-semibold text-white">24</p>
              <p className="text-xs text-slate-300">Active Services</p>
            </div>
            <div className="bg-white/10 backdrop-blur rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <Activity className="h-5 w-5 text-green-400" />
                <span className="text-xs text-green-400">+12%</span>
              </div>
              <p className="text-2xl font-semibold text-white">99.9%</p>
              <p className="text-xs text-slate-300">Uptime</p>
            </div>
            <div className="bg-white/10 backdrop-blur rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <Users className="h-5 w-5 text-purple-400" />
                <Badge variant="secondary" className="text-xs">Pro</Badge>
              </div>
              <p className="text-2xl font-semibold text-white">12</p>
              <p className="text-xs text-slate-300">Team Members</p>
            </div>
            <div className="bg-white/10 backdrop-blur rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <Shield className="h-5 w-5 text-yellow-400" />
                <CheckCircle2 className="h-4 w-4 text-green-400" />
              </div>
              <p className="text-2xl font-semibold text-white">256-bit</p>
              <p className="text-xs text-slate-300">Encryption</p>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Service Header"
        description="Header for service/resource management"
        
      >
        <div className="bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800">
          <div className="p-6">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-start gap-4">
                <IconBox icon={Terminal} color="blue" size="lg" />
                <div>
                  <div className="flex items-center gap-3 mb-1">
                    <h1 className="text-xl font-semibold text-slate-900 dark:text-white">api-gateway</h1>
                    <StatusBadge status="running" showPulse />
                    <Badge variant="outline">v2.4.1</Badge>
                  </div>
                  <p className="text-sm text-slate-600 dark:text-slate-400">
                    Main API gateway service handling all incoming requests
                  </p>
                  <div className="flex items-center gap-4 mt-3 text-sm text-slate-600 dark:text-slate-400">
                    <div className="flex items-center gap-1">
                      <GitBranch className="h-4 w-4" />
                      <span>main</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Clock className="h-4 w-4" />
                      <span>Deployed 3 hours ago</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Activity className="h-4 w-4" />
                      <span>2.4k req/min</span>
                    </div>
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm">
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Restart
                </Button>
                <Button variant="outline" size="sm">
                  <Pause className="h-4 w-4 mr-2" />
                  Pause
                </Button>
                <Button size="sm">
                  <Settings className="h-4 w-4 mr-2" />
                  Configure
                </Button>
                <Button variant="ghost" size="icon">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </div>
            </div>
            <div className="grid grid-cols-4 gap-4 pt-4 border-t border-slate-200 dark:border-slate-800">
              <div>
                <p className="text-xs text-slate-600 dark:text-slate-400 mb-1">CPU Usage</p>
                <div className="flex items-baseline gap-2">
                  <p className="text-lg font-semibold text-slate-900 dark:text-white">42%</p>
                  <p className="text-xs text-green-600">Normal</p>
                </div>
              </div>
              <div>
                <p className="text-xs text-slate-600 dark:text-slate-400 mb-1">Memory</p>
                <div className="flex items-baseline gap-2">
                  <p className="text-lg font-semibold text-slate-900 dark:text-white">1.2 GB</p>
                  <p className="text-xs text-slate-600 dark:text-slate-400">/ 4 GB</p>
                </div>
              </div>
              <div>
                <p className="text-xs text-slate-600 dark:text-slate-400 mb-1">Instances</p>
                <div className="flex items-baseline gap-2">
                  <p className="text-lg font-semibold text-slate-900 dark:text-white">3</p>
                  <p className="text-xs text-slate-600 dark:text-slate-400">replicas</p>
                </div>
              </div>
              <div>
                <p className="text-xs text-slate-600 dark:text-slate-400 mb-1">Health</p>
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 text-green-600" />
                  <p className="text-sm font-medium text-green-600">Healthy</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Header with Filters"
        description="Header with integrated filtering options"
        
      >
        <div className="bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800 p-6">
          <div className="flex items-center justify-between mb-4">
            <h1 className="text-xl font-semibold text-slate-900 dark:text-white">Services</h1>
            <Button size="sm">
              <Plus className="h-4 w-4 mr-2" />
              Add Service
            </Button>
          </div>
          <div className="flex items-center gap-3">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              <input
                type="text"
                placeholder="Search services..."
                className="w-full pl-10 pr-4 py-2 bg-slate-100 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <Button variant="outline" size="sm">
              <Filter className="h-4 w-4 mr-2" />
              Filter
            </Button>
            <Button variant="outline" size="sm">
              <SortDesc className="h-4 w-4 mr-2" />
              Sort
            </Button>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Alert Header"
        description="Header with system-wide alerts"
        
      >
        <div>
          <div className="bg-yellow-50 dark:bg-yellow-900/20 border-b border-yellow-200 dark:border-yellow-800 p-4">
            <div className="flex items-center gap-3">
              <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-400 shrink-0" />
              <div className="flex-1">
                <p className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                  Scheduled maintenance window
                </p>
                <p className="text-sm text-yellow-700 dark:text-yellow-300 mt-0.5">
                  System will be unavailable on Dec 15, 2:00 AM - 4:00 AM UTC for routine maintenance.
                </p>
              </div>
              <Button variant="ghost" size="sm" className="text-yellow-700 dark:text-yellow-300 hover:text-yellow-800 dark:hover:text-yellow-200">
                Learn more
              </Button>
            </div>
          </div>
          <div className="bg-white dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800 p-6">
            <h1 className="text-2xl font-semibold text-slate-900 dark:text-white">System Status</h1>
            <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
              Monitor the health and performance of all services
            </p>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Compact Header"
        description="Space-efficient header for nested views"
        
      >
        <div className="bg-slate-50 dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 px-6 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <h2 className="text-base font-medium text-slate-900 dark:text-white">Environment Variables</h2>
              <Badge variant="secondary">24 vars</Badge>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="ghost" size="sm">
                <Upload className="h-4 w-4 mr-2" />
                Import
              </Button>
              <Button variant="ghost" size="sm">
                <Plus className="h-4 w-4 mr-2" />
                Add Variable
              </Button>
            </div>
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}