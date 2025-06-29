import { Separator } from "@/components/atoms";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function SeparatorPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-foreground mb-4">
          Separator
        </h1>
        <p className="text-muted-foreground">
          Visual dividers for separating content sections.
        </p>
      </div>

      <ComponentShowcase
        title="Horizontal Separator"
        description="Default horizontal separator"
      >
        <div className="space-y-4">
          <div>
            <p className="text-sm text-muted-foreground mb-4">Content above separator</p>
            <Separator />
            <p className="text-sm text-muted-foreground mt-4">Content below separator</p>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Vertical Separator"
        description="Separator used between inline elements"
      >
        <div className="flex items-center space-x-4">
          <span>Item 1</span>
          <Separator orientation="vertical" className="h-4" />
          <span>Item 2</span>
          <Separator orientation="vertical" className="h-4" />
          <span>Item 3</span>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="With Text"
        description="Separator with centered text"
      >
        <div className="space-y-4">
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <Separator />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-white dark:bg-gray-900 px-2 text-muted-foreground">
                Or continue with
              </span>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Decorative Separators"
        description="Different styles and decorations"
      >
        <div className="space-y-6">
          <div>
            <p className="text-sm text-muted-foreground mb-2">Dashed separator</p>
            <hr className="border-t border-dashed border-border" />
          </div>
          
          <div>
            <p className="text-sm text-muted-foreground mb-2">Dotted separator</p>
            <hr className="border-t border-dotted border-border" />
          </div>
          
          <div>
            <p className="text-sm text-muted-foreground mb-2">Thick separator</p>
            <Separator className="h-0.5" />
          </div>
          
          <div>
            <p className="text-sm text-muted-foreground mb-2">Colored separator</p>
            <Separator className="bg-primary" />
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}