"use client";

import { StatDisplay } from "@/components/atoms";
import { TrendingUp, TrendingDown, Users, DollarSign, Activity, Package } from "lucide-react";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function StatDisplaysPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Stat Display Component
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Components for displaying statistics and metrics.
        </p>
      </div>

      <ComponentShowcase
        title="Dashboard Stats Cards"
        description="Stats as they appear in dashboards with card containers"
      >
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="rounded-lg border bg-card p-6">
            <StatDisplay label="Total Users" value="1,234" />
          </div>
          <div className="rounded-lg border bg-card p-6">
            <StatDisplay label="Revenue" value="$12,345" />
          </div>
          <div className="rounded-lg border bg-card p-6">
            <StatDisplay label="Growth Rate" value="+12.5%" className="text-success" />
          </div>
          <div className="rounded-lg border bg-card p-6">
            <StatDisplay label="Active Now" value="89" />
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Stats with Icons and Colors"
        description="Colored stat cards with icon indicators"
      >
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="rounded-lg border bg-card p-6 hover:shadow-md transition-shadow">
            <StatDisplay 
              label="Active Users" 
              value="892" 
              icon={Users}
              className="text-primary"
            />
          </div>
          <div className="rounded-lg border bg-card p-6 hover:shadow-md transition-shadow">
            <StatDisplay 
              label="Revenue" 
              value="$45,678" 
              icon={DollarSign}
              className="text-success"
            />
          </div>
          <div className="rounded-lg border bg-card p-6 hover:shadow-md transition-shadow">
            <StatDisplay 
              label="Orders" 
              value="156" 
              icon={Package}
              className="text-info"
            />
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Performance Metrics"
        description="Stats showing positive and negative trends"
      >
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="rounded-lg border border-success/20 bg-success/5 p-6">
            <StatDisplay 
              label="Sales Growth" 
              value="+23.5%" 
              icon={TrendingUp}
              className="text-success"
            />
          </div>
          <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-6">
            <StatDisplay 
              label="Bounce Rate" 
              value="42.3%" 
              icon={TrendingDown}
              className="text-destructive"
            />
          </div>
          <div className="rounded-lg border border-warning/20 bg-warning/5 p-6">
            <StatDisplay 
              label="CPU Usage" 
              value="78%" 
              icon={Activity}
              className="text-warning"
            />
          </div>
          <div className="rounded-lg border border-info/20 bg-info/5 p-6">
            <StatDisplay 
              label="Memory" 
              value="4.2 GB" 
              icon={Activity}
              className="text-info"
            />
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Stat Sizes Comparison"
        description="Same container size to compare stat component sizes"
      >
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="rounded-lg border-2 border-dashed bg-card p-8">
            <div className="text-xs text-muted-foreground mb-2">Small (size="sm")</div>
            <StatDisplay 
              label="Compact View" 
              value="123" 
              size="sm"
              icon={Users}
            />
          </div>
          <div className="rounded-lg border-2 border-dashed bg-card p-8">
            <div className="text-xs text-muted-foreground mb-2">Default (size="default")</div>
            <StatDisplay 
              label="Standard View" 
              value="456" 
              size="default"
              icon={Package}
            />
          </div>
          <div className="rounded-lg border-2 border-dashed bg-card p-8">
            <div className="text-xs text-muted-foreground mb-2">Large (size="lg")</div>
            <StatDisplay 
              label="Featured Metric" 
              value="789" 
              size="lg"
              icon={TrendingUp}
              className="text-primary"
            />
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Container Padding Variations"
        description="Same stat size with different container paddings"
      >
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="relative">
            <div className="absolute -top-2 left-2 text-xs font-medium text-muted-foreground bg-background px-1">p-2</div>
            <div className="rounded-lg border-2 border-dashed border-border bg-card/50 p-2">
              <div className="bg-muted/50 rounded">
                <StatDisplay label="Tight" value="42" size="sm" />
              </div>
            </div>
          </div>
          <div className="relative">
            <div className="absolute -top-2 left-2 text-xs font-medium text-muted-foreground bg-background px-1">p-4</div>
            <div className="rounded-lg border-2 border-dashed border-border bg-card/50 p-4">
              <div className="bg-muted/50 rounded">
                <StatDisplay label="Compact" value="42" size="sm" />
              </div>
            </div>
          </div>
          <div className="relative">
            <div className="absolute -top-2 left-2 text-xs font-medium text-muted-foreground bg-background px-1">p-6</div>
            <div className="rounded-lg border-2 border-dashed border-border bg-card/50 p-6">
              <div className="bg-muted/50 rounded">
                <StatDisplay label="Standard" value="42" size="sm" />
              </div>
            </div>
          </div>
          <div className="relative">
            <div className="absolute -top-2 left-2 text-xs font-medium text-muted-foreground bg-background px-1">p-8</div>
            <div className="rounded-lg border-2 border-dashed border-border bg-card/50 p-8">
              <div className="bg-muted/50 rounded">
                <StatDisplay label="Spacious" value="42" size="sm" />
              </div>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Minimal Stats Row"
        description="Compact stats for space-constrained layouts"
      >
        <div className="rounded-lg border bg-card p-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 divide-x">
            <StatDisplay label="Views" value="1.2K" size="sm" />
            <StatDisplay label="Likes" value="234" size="sm" className="pl-4" />
            <StatDisplay label="Comments" value="45" size="sm" className="pl-4" />
            <StatDisplay label="Shares" value="12" size="sm" className="pl-4" />
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}