import { Badge, Label, Caption } from "@/components/atoms";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function BadgesPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-foreground mb-4">
          Badges & Labels
        </h1>
        <p className="text-muted-foreground">
          Components for displaying status, labels, and metadata.
        </p>
      </div>

      <ComponentShowcase
        title="Badge Variants"
        description="Different badge styles for various contexts"
      >
        <div className="flex flex-wrap gap-2">
          <Badge>Default</Badge>
          <Badge variant="secondary">Secondary</Badge>
          <Badge variant="destructive">Destructive</Badge>
          <Badge variant="outline">Outline</Badge>
          <Badge variant="success">Success</Badge>
          <Badge variant="info">Info</Badge>
          <Badge variant="warning">Warning</Badge>
          <Badge variant="error">Error</Badge>
          <Badge variant="purple">Purple</Badge>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Badge Sizes"
        description="Available badge sizes"
      >
        <div className="flex flex-wrap items-center gap-2">
          <Badge size="small">Small</Badge>
          <Badge size="default">Default</Badge>
          <Badge size="large">Large</Badge>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Labels"
        description="Form labels for inputs and fields"
      >
        <div className="space-y-4 max-w-md">
          <div>
            <Label>Default Label</Label>
            <p className="text-sm text-muted-foreground mt-1">
              Standard label for form fields
            </p>
          </div>
          
          <div>
            <Label htmlFor="required-field">
              Required Field <span className="text-red-500">*</span>
            </Label>
            <p className="text-sm text-muted-foreground mt-1">
              Label with required indicator
            </p>
          </div>

          <div>
            <Label className="text-xs">Small Label</Label>
            <p className="text-sm text-muted-foreground mt-1">
              Smaller label variant
            </p>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Captions"
        description="Helper text and error messages"
      >
        <div className="space-y-4 max-w-md">
          <div>
            <Label>Field Label</Label>
            <Caption>This is a helpful description</Caption>
          </div>
          
          <div>
            <Label>Field with Error</Label>
            <Caption error>This field has an error</Caption>
          </div>

          <div>
            <Label>Field with Success</Label>
            <Caption className="text-success">
              ✓ Field validated successfully
            </Caption>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Combined Examples"
        description="Badges used in different contexts"
      >
        <div className="space-y-4">
          <div className="flex items-center gap-2">
            <span className="text-sm">User Status:</span>
            <Badge variant="success" size="small">Active</Badge>
            <Badge variant="warning" size="small">Pending</Badge>
            <Badge variant="error" size="small">Inactive</Badge>
          </div>

          <div className="flex items-center gap-2">
            <span className="text-sm">Environment:</span>
            <Badge variant="info">Development</Badge>
            <Badge variant="warning">Staging</Badge>
            <Badge variant="success">Production</Badge>
          </div>

          <div className="flex items-center gap-2">
            <span className="text-sm">Permissions:</span>
            <Badge variant="purple">Admin</Badge>
            <Badge variant="secondary">Developer</Badge>
            <Badge variant="outline">Viewer</Badge>
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}