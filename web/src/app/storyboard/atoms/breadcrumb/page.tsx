import { Breadcrumb } from "@/components/atoms";
import { ChevronRight, Home, Settings, Users } from "lucide-react";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function BreadcrumbPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-foreground mb-4">
          Breadcrumb
        </h1>
        <p className="text-muted-foreground">
          Navigation breadcrumbs for showing page hierarchy.
        </p>
      </div>

      <ComponentShowcase
        title="Basic Breadcrumb"
        description="Standard breadcrumb navigation"
      >
        <div className="space-y-4">
          <Breadcrumb
            items={[
              { label: "Home", href: "/" },
              { label: "Dashboard", href: "/dashboard" },
              { label: "Settings" },
            ]}
          />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="With Icons"
        description="Breadcrumb with icon support"
      >
        <div className="space-y-4">
          <Breadcrumb
            items={[
              { label: "Home", href: "/", icon: Home },
              { label: "Team", href: "/team", icon: Users },
              { label: "Settings", icon: Settings },
            ]}
          />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Custom Separator"
        description="Breadcrumb with custom separator"
      >
        <div className="space-y-4">
          <Breadcrumb
            items={[
              { label: "Projects", href: "/" },
              { label: "Kloudlite", href: "/projects/kloudlite" },
              { label: "Environments", href: "/projects/kloudlite/environments" },
              { label: "Production" },
            ]}
            separator="/"
          />
          
          <Breadcrumb
            items={[
              { label: "Home", href: "/" },
              { label: "Products", href: "/products" },
              { label: "Category", href: "/products/category" },
              { label: "Item" },
            ]}
            separator="•"
          />
          
          <Breadcrumb
            items={[
              { label: "Step 1", href: "/" },
              { label: "Step 2", href: "/step2" },
              { label: "Step 3", href: "/step3" },
              { label: "Complete" },
            ]}
            separator="→"
          />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Sizes"
        description="Different breadcrumb sizes"
      >
        <div className="space-y-4">
          <Breadcrumb
            size="sm"
            items={[
              { label: "Home", href: "/" },
              { label: "Products", href: "/products" },
              { label: "Details" },
            ]}
          />
          <Breadcrumb
            size="default"
            items={[
              { label: "Home", href: "/" },
              { label: "Products", href: "/products" },
              { label: "Details" },
            ]}
          />
          <Breadcrumb
            size="lg"
            items={[
              { label: "Home", href: "/" },
              { label: "Products", href: "/products" },
              { label: "Details" },
            ]}
          />
        </div>
      </ComponentShowcase>
    </div>
  );
}