import { BaseCard, CardHeader, CardContent, CardTitle, CardDescription, StatusBadge } from "@/components/molecules";
import { ServiceCard, ResourceInfoCard, StatsCard, QuickActionCard } from "@/components/molecules";
import { Button, Badge, IconBox } from "@/components/atoms";
import { Database, FileJson, GitBranch, Settings, Terminal, TrendingUp, Users } from "lucide-react";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function CardsPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Card Components
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Various card components for displaying grouped content.
        </p>
      </div>

      <ComponentShowcase
        title="Base Card"
        description="Foundation card component with different padding options"
      >
        <div className="grid gap-4">
          <BaseCard>
            <p>Default padding card content</p>
          </BaseCard>
          
          <BaseCard padding="sm">
            <p>Small padding card content</p>
          </BaseCard>
          
          <BaseCard padding="lg">
            <p>Large padding card content</p>
          </BaseCard>

          <BaseCard padding="none">
            <p>No padding card content</p>
          </BaseCard>
          
          <BaseCard >
            <CardHeader>
              <CardTitle>Card with Header</CardTitle>
              <CardDescription>And a description</CardDescription>
            </CardHeader>
            <CardContent>
              <p>Card content goes here</p>
            </CardContent>
          </BaseCard>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Stats Card"
        description="Cards for displaying statistics with trends"
      >
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          <StatsCard
            label="Active Users"
            value="1,234"
            subValue="of 2,000 total"
            trend={{ value: 12.5, isPositive: true }}
            icon={Users}
            iconColor="text-blue-600 dark:text-blue-400"
            bgColor="bg-blue-50 dark:bg-blue-500/10"
          />
          
          <StatsCard
            label="Revenue"
            value="$45.2K"
            subValue="this month"
            trend={{ value: 8.3, isPositive: false }}
            icon={TrendingUp}
            iconColor="text-green-600 dark:text-green-400"
            bgColor="bg-green-50 dark:bg-green-500/10"
          />
          
          <StatsCard
            label="Performance"
            value="98.5%"
            subValue="uptime"
            icon={Settings}
            iconColor="text-purple-600 dark:text-purple-400"
            bgColor="bg-purple-50 dark:bg-purple-500/10"
          />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Quick Action Cards"
        description="Interactive cards for quick actions"
      >
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          <QuickActionCard
            href="#"
            icon={GitBranch}
            iconColor="text-blue-600 dark:text-blue-400"
            bgColor="bg-blue-50 dark:bg-blue-500/10"
            title="Create Environment"
            description="Spin up a new dev environment"
          />
          
          <QuickActionCard
            href="#"
            icon={Terminal}
            iconColor="text-green-600 dark:text-green-400"
            bgColor="bg-green-50 dark:bg-green-500/10"
            title="Start Workspace"
            description="Launch your dev workspace"
          />
          
          <QuickActionCard
            href="#"
            icon={Database}
            iconColor="text-purple-600 dark:text-purple-400"
            bgColor="bg-purple-50 dark:bg-purple-500/10"
            title="Deploy Service"
            description="Add databases & services"
          />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Service Card"
        description="Cards for displaying service information"
      >
        <ServiceCard
          id="1"
          name="primary-mongodb"
          type="mongodb"
          version="7.0"
          status="active"
          icon={Database}
          iconColor="green"
          createdAt="Jan 15, 2024"
          lastModified="Jun 25, 2024"
          onDelete={() => console.log("Delete service")}
        >
          <ResourceInfoCard
            icon={FileJson}
            title="Resources"
            description="3 databases, 24 collections"
            href="#"
          />
          
          <ResourceInfoCard
            icon={Settings}
            title="Configurations"
            description="12 configurations • Last export Jun 24, 2024"
            href="#"
          />
        </ServiceCard>
      </ComponentShowcase>

      <ComponentShowcase
        title="Interactive Cards"
        description="Cards with hover effects and interactions"
      >
        <div className="grid md:grid-cols-2 gap-4">
          <BaseCard interactive className="cursor-pointer">
            <h3 className="font-semibold mb-2">Interactive Card</h3>
            <p className="text-sm text-slate-600 dark:text-slate-400">
              This card has hover effects and can be clicked
            </p>
          </BaseCard>
          
          <BaseCard className="border-2 border-blue-500">
            <h3 className="font-semibold mb-2">Custom Border</h3>
            <p className="text-sm text-slate-600 dark:text-slate-400">
              Card with custom border styling
            </p>
          </BaseCard>
        </div>
      </ComponentShowcase>
    </div>
  );
}