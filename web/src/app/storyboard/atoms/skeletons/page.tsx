"use client";

import { Skeleton } from "@/components/atoms";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function SkeletonsPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Skeleton Component
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Loading placeholders that show while content is being fetched.
        </p>
      </div>

      <ComponentShowcase
        title="Basic Skeletons"
        description="Different skeleton shapes and sizes"
      >
        <div className="space-y-4">
          <div className="space-y-2">
            <p className="text-sm text-muted-foreground">Text skeleton</p>
            <Skeleton className="h-4 w-[250px]" />
            <Skeleton className="h-4 w-[200px]" />
            <Skeleton className="h-4 w-[150px]" />
          </div>

          <div className="space-y-2">
            <p className="text-sm text-muted-foreground">Button skeleton</p>
            <Skeleton className="h-9 w-[100px] rounded-md" />
          </div>

          <div className="space-y-2">
            <p className="text-sm text-muted-foreground">Avatar skeleton</p>
            <Skeleton className="h-12 w-12 rounded-full" />
          </div>

          <div className="space-y-2">
            <p className="text-sm text-muted-foreground">Input skeleton</p>
            <Skeleton className="h-9 w-full rounded-md" />
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Card Skeleton"
        description="Complex skeleton for card layouts"
      >
        <div className="w-full max-w-md">
          <div className="rounded-lg border bg-card p-6 space-y-4">
            <div className="flex items-center space-x-4">
              <Skeleton className="h-12 w-12 rounded-full" />
              <div className="space-y-2">
                <Skeleton className="h-4 w-[200px]" />
                <Skeleton className="h-3 w-[150px]" />
              </div>
            </div>
            <div className="space-y-2">
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-3/4" />
            </div>
            <div className="flex space-x-2">
              <Skeleton className="h-8 w-20 rounded-md" />
              <Skeleton className="h-8 w-20 rounded-md" />
            </div>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Table Skeleton"
        description="Skeleton for table layouts"
      >
        <div className="w-full">
          <div className="space-y-2">
            {/* Header */}
            <div className="flex space-x-4 pb-2 border-b">
              <Skeleton className="h-4 w-[100px]" />
              <Skeleton className="h-4 w-[150px]" />
              <Skeleton className="h-4 w-[100px]" />
              <Skeleton className="h-4 w-[80px]" />
            </div>
            {/* Rows */}
            {[...Array(5)].map((_, i) => (
              <div key={i} className="flex space-x-4 py-2">
                <Skeleton className="h-4 w-[100px]" />
                <Skeleton className="h-4 w-[150px]" />
                <Skeleton className="h-4 w-[100px]" />
                <Skeleton className="h-4 w-[80px]" />
              </div>
            ))}
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Pulse Animation"
        description="Skeleton with pulse animation"
      >
        <div className="space-y-4">
          <div className="space-y-2">
            <p className="text-sm text-muted-foreground">Default animation</p>
            <Skeleton className="h-20 w-full rounded-md" />
          </div>
          
          <div className="space-y-2">
            <p className="text-sm text-muted-foreground">Custom styling</p>
            <Skeleton className="h-20 w-full rounded-md bg-primary/10" />
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}