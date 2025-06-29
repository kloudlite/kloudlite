import { Avatar } from "@/components/atoms";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function AvatarsPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-foreground mb-4">
          Avatars
        </h1>
        <p className="text-muted-foreground">
          User avatar components for displaying profile images or initials.
        </p>
      </div>

      <ComponentShowcase
        title="Avatar Sizes"
        description="Available avatar sizes"
      >
        <div className="flex flex-wrap items-center gap-4">
          <Avatar size="xs" fallback="XS" />
          <Avatar size="sm" fallback="SM" />
          <Avatar size="md" fallback="MD" />
          <Avatar size="lg" fallback="LG" />
          <Avatar size="xl" fallback="XL" />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Avatar Colors"
        description="Different color variants for avatars"
      >
        <div className="flex flex-wrap items-center gap-4">
          <Avatar color="slate" fallback="SL" />
          <Avatar color="blue" fallback="BL" />
          <Avatar color="green" fallback="GR" />
          <Avatar color="yellow" fallback="YL" />
          <Avatar color="red" fallback="RD" />
          <Avatar color="purple" fallback="PR" />
          <Avatar color="gradient" fallback="GD" />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="With Images"
        description="Avatars with profile images"
      >
        <div className="flex flex-wrap items-center gap-4">
          <Avatar 
            size="sm"
            src="https://api.dicebear.com/7.x/avataaars/svg?seed=John"
            alt="John Doe"
            fallback="JD"
          />
          <Avatar 
            size="md"
            src="https://api.dicebear.com/7.x/avataaars/svg?seed=Sarah"
            alt="Sarah Chen"
            fallback="SC"
          />
          <Avatar 
            size="lg"
            src="https://api.dicebear.com/7.x/avataaars/svg?seed=Michael"
            alt="Michael Rodriguez"
            fallback="MR"
          />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Fallback Examples"
        description="Avatars with initials when no image is provided"
      >
        <div className="flex flex-wrap items-center gap-4">
          <div className="text-center">
            <Avatar fallback="JD" className="mb-2" />
            <p className="text-xs text-muted-foreground">John Doe</p>
          </div>
          <div className="text-center">
            <Avatar fallback="SC" color="blue" className="mb-2" />
            <p className="text-xs text-muted-foreground">Sarah Chen</p>
          </div>
          <div className="text-center">
            <Avatar fallback="MR" color="green" className="mb-2" />
            <p className="text-xs text-muted-foreground">Michael R.</p>
          </div>
          <div className="text-center">
            <Avatar fallback="K" color="purple" className="mb-2" />
            <p className="text-xs text-muted-foreground">Karthik</p>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Avatar Groups"
        description="Multiple avatars displayed together"
      >
        <div className="space-y-4">
          <div className="flex -space-x-2">
            <Avatar size="sm" fallback="JD" className="ring-2 ring-card hover:z-10" />
            <Avatar size="sm" fallback="SC" color="blue" className="ring-2 ring-card hover:z-10" />
            <Avatar size="sm" fallback="MR" color="green" className="ring-2 ring-card hover:z-10" />
            <Avatar size="sm" fallback="+3" color="slate" className="ring-2 ring-card hover:z-10" />
          </div>

          <div className="flex -space-x-3">
            <Avatar fallback="PT" color="gradient" className="ring-2 ring-card hover:z-10" />
            <Avatar fallback="EW" color="purple" className="ring-2 ring-card hover:z-10" />
            <Avatar fallback="JW" color="yellow" className="ring-2 ring-card hover:z-10" />
            <Avatar fallback="AT" color="red" className="ring-2 ring-card hover:z-10" />
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}