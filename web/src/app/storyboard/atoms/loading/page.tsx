"use client";

import { Spinner, Progress, Button } from "@/components/atoms";
import { ComponentShowcase } from "../../_components/component-showcase";
import { useState } from "react";

export default function LoadingPage() {
  const [progress, setProgress] = useState(60);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-foreground mb-4">
          Loading & Progress
        </h1>
        <p className="text-muted-foreground">
          Loading indicators and progress bars for async operations.
        </p>
      </div>

      <ComponentShowcase
        title="Spinner Sizes"
        description="Different spinner sizes"
      >
        <div className="flex items-center gap-4">
          <Spinner size="xs" />
          <Spinner size="sm" />
          <Spinner size="md" />
          <Spinner size="lg" />
          <Spinner size="xl" />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Spinner Colors"
        description="Various spinner color options"
      >
        <div className="flex items-center gap-4">
          <Spinner color="default" />
          <Spinner color="primary" />
          <Spinner color="secondary" />
          <Spinner color="success" />
          <Spinner color="warning" />
          <Spinner color="destructive" />
          <Spinner color="info" />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Loading States"
        description="Common loading patterns"
      >
        <div className="space-y-4">
          <div className="flex items-center gap-2">
            <Spinner size="sm" />
            <span className="text-sm text-muted-foreground">Loading...</span>
          </div>

          <Button className="w-32 pointer-events-none" tabIndex={-1}>
            <Spinner size="sm" className="mr-2 text-primary-foreground" />
            Loading
          </Button>

          <div className="flex flex-col items-center justify-center p-8 border border-dashed rounded-lg">
            <Spinner size="lg" className="mb-4" />
            <p className="text-sm text-muted-foreground">Loading content...</p>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Progress Bar"
        description="Basic progress indicators"
      >
        <div className="space-y-4 max-w-md">
          <Progress value={30} />
          <Progress value={60} color="success" />
          <Progress value={90} color="warning" />
          <Progress value={100} color="info" />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Progress Sizes"
        description="Different progress bar sizes"
      >
        <div className="space-y-4 max-w-md">
          <Progress value={40} size="sm" />
          <Progress value={40} size="md" />
          <Progress value={40} size="lg" />
          <Progress value={40} size="xl" />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Progress with Label"
        description="Progress bars with percentage labels"
      >
        <div className="space-y-4 max-w-md">
          <Progress value={25} showLabel />
          <Progress value={75} color="success" showLabel />
          
          <div className="space-y-2">
            <Progress value={progress} color="primary" showLabel />
            <div className="flex gap-2 items-center">
              <Button 
                size="sm" 
                onClick={() => setProgress(Math.max(0, progress - 10))}
              >
                -10%
              </Button>
              <Button 
                size="sm" 
                onClick={() => setProgress(Math.min(100, progress + 10))}
              >
                +10%
              </Button>
              <span className="text-sm text-muted-foreground ml-2">
                (Current: {progress}%)
              </span>
            </div>
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}