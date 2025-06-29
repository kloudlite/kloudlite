"use client";

import { ComponentShowcase } from "../../_components/component-showcase";
import { Heading, Text } from "@/components/atoms";

export default function TypographyPage() {
  return (
    <div className="space-y-8">
      <div>
        <Heading level={2} className="mb-4">Typography</Heading>
        <Text color="secondary">
          Text styles and typography components using design system tokens.
        </Text>
      </div>

      <ComponentShowcase
        title="Headings"
        description="All heading levels with design system tokens"
      >
        <div className="space-y-4">
          <Heading level={1}>Heading 1 (48px)</Heading>
          <Heading level={2}>Heading 2 (36px)</Heading>
          <Heading level={3}>Heading 3 (30px)</Heading>
          <Heading level={4}>Heading 4 (24px)</Heading>
          <Heading level={5}>Heading 5 (20px)</Heading>
          <Heading level={6}>Heading 6 (18px)</Heading>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Text Sizes"
        description="All available text sizes from design tokens"
      >
        <div className="space-y-3">
          <Text size="6xl">6xl - The quick brown fox (60px)</Text>
          <Text size="5xl">5xl - The quick brown fox (48px)</Text>
          <Text size="4xl">4xl - The quick brown fox (36px)</Text>
          <Text size="3xl">3xl - The quick brown fox (30px)</Text>
          <Text size="2xl">2xl - The quick brown fox (24px)</Text>
          <Text size="xl">xl - The quick brown fox jumps over the lazy dog (20px)</Text>
          <Text size="lg">lg - The quick brown fox jumps over the lazy dog (18px)</Text>
          <Text size="base">base - The quick brown fox jumps over the lazy dog (16px)</Text>
          <Text size="sm">sm - The quick brown fox jumps over the lazy dog (14px)</Text>
          <Text size="xs">xs - The quick brown fox jumps over the lazy dog (12px)</Text>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Font Weights"
        description="Available font weight tokens"
      >
        <div className="space-y-2">
          <Text weight="light">Light weight (300)</Text>
          <Text weight="normal">Normal weight (400)</Text>
          <Text weight="medium">Medium weight (500)</Text>
          <Text weight="semibold">Semibold weight (600)</Text>
          <Text weight="bold">Bold weight (700)</Text>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Text Colors"
        description="Semantic color tokens for text"
      >
        <div className="space-y-2">
          <Text color="default">Default text color</Text>
          <Text color="secondary">Secondary text color</Text>
          <Text color="muted">Muted text color</Text>
          <Text color="primary">Primary brand color</Text>
          <Text color="success">Success status color</Text>
          <Text color="warning">Warning status color</Text>
          <Text color="error">Error status color</Text>
          <Text color="info">Info status color</Text>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Letter Spacing"
        description="Typography tracking options"
      >
        <div className="space-y-2">
          <Text tracking="tighter" size="lg">Tighter letter spacing (-0.05em)</Text>
          <Text tracking="tight" size="lg">Tight letter spacing (-0.025em)</Text>
          <Text tracking="normal" size="lg">Normal letter spacing (0em)</Text>
          <Text tracking="wide" size="lg">Wide letter spacing (0.025em)</Text>
          <Text tracking="wider" size="lg">Wider letter spacing (0.05em)</Text>
          <Text tracking="widest" size="lg">Widest letter spacing (0.1em)</Text>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Line Heights"
        description="Text leading options"
      >
        <div className="space-y-4">
          <div>
            <Text weight="medium" className="mb-1">Tight Line Height</Text>
            <Text leading="tight" className="max-w-md">
              Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
            </Text>
          </div>
          <div>
            <Text weight="medium" className="mb-1">Normal Line Height</Text>
            <Text leading="normal" className="max-w-md">
              Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
            </Text>
          </div>
          <div>
            <Text weight="medium" className="mb-1">Relaxed Line Height</Text>
            <Text leading="relaxed" className="max-w-md">
              Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
            </Text>
          </div>
          <div>
            <Text weight="medium" className="mb-1">Loose Line Height</Text>
            <Text leading="loose" className="max-w-md">
              Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
            </Text>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Text Alignment"
        description="Paragraph alignment options"
      >
        <div className="space-y-4">
          <Text align="left" className="p-4 bg-slate-50 dark:bg-slate-900 rounded">
            Left aligned text (default)
          </Text>
          <Text align="center" className="p-4 bg-slate-50 dark:bg-slate-900 rounded">
            Center aligned text
          </Text>
          <Text align="right" className="p-4 bg-slate-50 dark:bg-slate-900 rounded">
            Right aligned text
          </Text>
          <Text align="justify" className="p-4 bg-slate-50 dark:bg-slate-900 rounded">
            Justified text alignment spreads the text evenly across the full width of the container, creating clean edges on both sides.
          </Text>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Monospace Font"
        description="Code and technical content"
      >
        <div className="space-y-2">
          <Text font="mono">const greeting = "Hello, World!";</Text>
          <Text font="mono" size="sm" color="muted">// This is a code comment</Text>
          <Text font="mono" className="p-3 bg-slate-100 dark:bg-slate-900 rounded">
            npm install @kloudlite/design-system
          </Text>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Custom Heading Sizes"
        description="Using size prop instead of level"
      >
        <div className="space-y-3">
          <Heading size="6xl" weight="bold">Custom 6xl Heading</Heading>
          <Heading size="3xl" weight="medium" color="primary">3xl Primary Heading</Heading>
          <Heading size="xl" weight="normal" color="secondary">xl Secondary Heading</Heading>
          <Heading size="base" weight="semibold" color="muted">Base Size Heading</Heading>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Practical Examples"
        description="Common typography patterns"
      >
        <div className="space-y-6">
          {/* Card Example */}
          <div className="p-6 border rounded-lg">
            <Heading level={4} className="mb-2">Card Title</Heading>
            <Text color="secondary" className="mb-4">
              This is a description that provides context about the card content.
            </Text>
            <Text size="sm" color="muted">
              Last updated 2 hours ago
            </Text>
          </div>

          {/* Section Example */}
          <div>
            <Heading level={3} className="mb-1">Section Header</Heading>
            <Text color="secondary" size="lg" className="mb-4">
              A brief introduction to this section
            </Text>
            <Text className="mb-2">
              This is the main body text of the section. It uses the default text size and color for optimal readability.
            </Text>
            <Text size="sm" color="muted">
              Additional notes or metadata can use smaller, muted text.
            </Text>
          </div>

          {/* Alert Example */}
          <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
            <Heading level={6} color="warning" className="mb-1">
              Warning
            </Heading>
            <Text size="sm" color="warning">
              Please review your configuration before proceeding.
            </Text>
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}