"use client";

import { Button } from "@/components/atoms";
import { 
  ArrowRight, 
  Check, 
  Download, 
  Plus, 
  RefreshCw, 
  Settings,
  Trash2
} from "lucide-react";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function ButtonsPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Buttons
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Button components for triggering actions and navigation.
        </p>
      </div>

      <ComponentShowcase
        title="Button Variants"
        description="Different button styles for various use cases"
      >
        <div className="flex flex-wrap gap-4">
          <Button>Default</Button>
          <Button variant="destructive">Destructive</Button>
          <Button variant="outline">Outline</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="link">Link</Button>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Button Sizes"
        description="Available button sizes"
      >
        <div className="flex flex-wrap items-center gap-4">
          <Button size="sm">Small</Button>
          <Button size="default">Default</Button>
          <Button size="lg">Large</Button>
          <Button size="icon">
            <Settings className="h-4 w-4" />
          </Button>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="With Icons"
        description="Buttons with leading or trailing icons"
      >
        <div className="flex flex-wrap gap-4">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            Create New
          </Button>
          <Button variant="outline">
            <Download className="mr-2 h-4 w-4" />
            Download
          </Button>
          <Button variant="secondary">
            Continue
            <ArrowRight className="ml-2 h-4 w-4" />
          </Button>
          <Button variant="destructive">
            <Trash2 className="mr-2 h-4 w-4" />
            Delete
          </Button>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Button States"
        description="Different button states"
      >
        <div className="flex flex-wrap gap-4">
          <Button>Normal</Button>
          <Button disabled>Disabled</Button>
          <Button>
            <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
            Loading...
          </Button>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Full Width"
        description="Buttons that span the full width of their container"
      >
        <div className="max-w-sm space-y-2">
          <Button className="w-full">Full Width Button</Button>
          <Button variant="outline" className="w-full">
            <Check className="mr-2 h-4 w-4" />
            Full Width with Icon
          </Button>
        </div>
      </ComponentShowcase>
    </div>
  );
}