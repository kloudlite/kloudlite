import { Tooltip, Button, Badge, Avatar } from "@/components/atoms";
import { Info, HelpCircle, Settings } from "lucide-react";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function TooltipPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-foreground mb-4">
          Tooltip
        </h1>
        <p className="text-muted-foreground">
          Contextual information displayed on hover or focus.
        </p>
      </div>

      <ComponentShowcase
        title="Basic Tooltip"
        description="Simple tooltip on various elements"
      >
        <div className="flex items-center gap-4">
          <Tooltip content="This is a tooltip">
            <Button variant="outline">Hover me</Button>
          </Tooltip>

          <Tooltip content="Edit settings">
            <Button variant="ghost" size="icon">
              <Settings className="h-4 w-4" />
            </Button>
          </Tooltip>

          <Tooltip content="Click for more information">
            <HelpCircle className="h-5 w-5 text-muted-foreground cursor-help" />
          </Tooltip>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Tooltip Positions"
        description="Tooltips in different positions"
      >
        <div className="flex items-center justify-center gap-8 py-8">
          <Tooltip content="Top tooltip" side="top">
            <Button variant="outline">Top</Button>
          </Tooltip>

          <Tooltip content="Right tooltip" side="right">
            <Button variant="outline">Right</Button>
          </Tooltip>

          <Tooltip content="Bottom tooltip" side="bottom">
            <Button variant="outline">Bottom</Button>
          </Tooltip>

          <Tooltip content="Left tooltip" side="left">
            <Button variant="outline">Left</Button>
          </Tooltip>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Tooltip Variants"
        description="Different tooltip styles"
      >
        <div className="flex items-center gap-4">
          <Tooltip content="Default tooltip" variant="default">
            <Button variant="outline">Default</Button>
          </Tooltip>

          <Tooltip content="Secondary tooltip" variant="secondary">
            <Button variant="outline">Secondary</Button>
          </Tooltip>

          <Tooltip content="Destructive action!" variant="destructive">
            <Button variant="outline">Destructive</Button>
          </Tooltip>

          <Tooltip content="Dark tooltip with shadow" variant="dark">
            <Button variant="outline">Dark</Button>
          </Tooltip>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Complex Content"
        description="Tooltips with rich content"
      >
        <div className="flex items-center gap-4">
          <Tooltip 
            content={
              <div className="space-y-2">
                <p className="font-semibold">User Information</p>
                <p className="text-xs">Name: John Doe</p>
                <p className="text-xs">Role: Administrator</p>
                <p className="text-xs">Last login: 2 hours ago</p>
              </div>
            }
            variant="dark"
          >
            <Avatar fallback="JD" />
          </Tooltip>

          <Tooltip 
            content={
              <div className="flex items-center gap-2">
                <Info className="h-4 w-4" />
                <span>Press Cmd+K to open command menu</span>
              </div>
            }
          >
            <Badge variant="secondary">Shortcut</Badge>
          </Tooltip>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Delay Duration"
        description="Control tooltip show delay"
      >
        <div className="flex items-center gap-4">
          <Tooltip content="Instant (0ms delay)" delayDuration={0}>
            <Button variant="outline">Instant</Button>
          </Tooltip>

          <Tooltip content="Fast (100ms delay)" delayDuration={100}>
            <Button variant="outline">Fast</Button>
          </Tooltip>

          <Tooltip content="Default (200ms delay)" delayDuration={200}>
            <Button variant="outline">Default</Button>
          </Tooltip>

          <Tooltip content="Slow (500ms delay)" delayDuration={500}>
            <Button variant="outline">Slow</Button>
          </Tooltip>
        </div>
      </ComponentShowcase>
    </div>
  );
}