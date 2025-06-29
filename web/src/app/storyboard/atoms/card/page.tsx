import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle, Button } from "@/components/atoms";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function CardPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-foreground mb-4">
          Card
        </h1>
        <p className="text-muted-foreground">
          Container components for grouping related content.
        </p>
      </div>

      <ComponentShowcase
        title="Card with Padding"
        description="Cards with different padding options"
      >
        <div className="space-y-4 max-w-md">
          <Card padding>
            <p>This is a card with default padding (p-6).</p>
          </Card>
          
          <Card padding="p-4">
            <p>This is a card with custom padding (p-4).</p>
          </Card>
          
          <Card padding="p-8">
            <p>This is a card with larger padding (p-8).</p>
          </Card>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Card with Header"
        description="Card with title and description"
      >
        <div className="max-w-md">
          <Card>
            <CardHeader className="p-6 pb-4">
              <CardTitle>Card Title</CardTitle>
              <CardDescription>
                This is a description that provides more context about the card content.
              </CardDescription>
            </CardHeader>
            <CardContent className="px-6 pb-6">
              <p>The main content of the card goes here.</p>
            </CardContent>
          </Card>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Card with Footer"
        description="Complete card with header, content, and footer"
      >
        <div className="max-w-md">
          <Card>
            <CardHeader className="p-6 pb-4">
              <CardTitle>Create New Project</CardTitle>
              <CardDescription>
                Deploy your new project in one-click.
              </CardDescription>
            </CardHeader>
            <CardContent className="px-6 pb-4">
              <p className="text-sm text-muted-foreground">
                Set up your project by configuring the necessary settings and choosing your deployment options.
              </p>
            </CardContent>
            <CardFooter className="px-6 pb-6 flex justify-between">
              <Button variant="outline">Cancel</Button>
              <Button>Deploy</Button>
            </CardFooter>
          </Card>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Interactive Card"
        description="Cards with hover effects"
      >
        <div className="grid md:grid-cols-3 gap-4">
          <Card padding className="cursor-pointer hover:shadow-lg transition-shadow">
            <h3 className="font-semibold text-lg mb-2">Hover Effect</h3>
            <p className="text-sm text-muted-foreground">This card has a hover shadow effect.</p>
          </Card>

          <Card padding className="cursor-pointer border-2 hover:border-primary transition-colors">
            <h3 className="font-semibold text-lg mb-2">Border Highlight</h3>
            <p className="text-sm text-muted-foreground">This card highlights its border on hover.</p>
          </Card>

          <Card padding className="cursor-pointer hover:bg-accent transition-colors">
            <h3 className="font-semibold text-lg mb-2">Background Change</h3>
            <p className="text-sm text-muted-foreground">This card changes background on hover.</p>
          </Card>
        </div>
      </ComponentShowcase>
    </div>
  );
}