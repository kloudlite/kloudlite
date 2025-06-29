import { BaseCard } from "@/components/molecules";
import { Circle, Square, Triangle, Layout } from "lucide-react";

export default function StoryboardOverview() {
  return (
    <div className="max-w-6xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-slate-900 dark:text-white mb-4">
          Kloudlite Component Storyboard
        </h1>
        <p className="text-lg text-slate-600 dark:text-slate-400">
          A comprehensive showcase of all components in the Kloudlite design system, organized following the Atomic Design methodology.
        </p>
      </div>

      <div className="grid md:grid-cols-2 gap-6 mb-12">
        <BaseCard className="border-slate-200 dark:border-slate-800">
          <h3 className="flex items-center gap-3 text-lg font-semibold text-slate-900 dark:text-white mb-4">
            <Circle className="h-5 w-5 text-blue-600" />
            Atoms
          </h3>
          <p className="text-sm text-slate-600 dark:text-slate-400 mb-4">
            The foundational building blocks of our design system. These are the smallest, most basic components that cannot be broken down further.
          </p>
          <ul className="text-sm space-y-1 text-slate-600 dark:text-slate-400">
            <li>• Buttons, Inputs, Labels</li>
            <li>• Typography elements</li>
            <li>• Badges, Icons, Avatars</li>
          </ul>
        </BaseCard>

        <BaseCard className="border-slate-200 dark:border-slate-800">
          <h3 className="flex items-center gap-3 text-lg font-semibold text-slate-900 dark:text-white mb-4">
            <Triangle className="h-5 w-5 text-green-600" />
            Molecules
          </h3>
          <p className="text-sm text-slate-600 dark:text-slate-400 mb-4">
            Groups of atoms bonded together to form relatively simple UI components with a single purpose.
          </p>
          <ul className="text-sm space-y-1 text-slate-600 dark:text-slate-400">
            <li>• Cards, Form fields</li>
            <li>• Dialogs, Dropdowns</li>
            <li>• Status indicators</li>
          </ul>
        </BaseCard>

        <BaseCard className="border-slate-200 dark:border-slate-800">
          <h3 className="flex items-center gap-3 text-lg font-semibold text-slate-900 dark:text-white mb-4">
            <Square className="h-5 w-5 text-purple-600" />
            Organisms
          </h3>
          <p className="text-sm text-slate-600 dark:text-slate-400 mb-4">
            Complex components composed of groups of molecules and/or atoms working together as a unit.
          </p>
          <ul className="text-sm space-y-1 text-slate-600 dark:text-slate-400">
            <li>• Navigation components</li>
            <li>• Headers and Footers</li>
            <li>• Tables and Lists</li>
          </ul>
        </BaseCard>

        <BaseCard className="border-slate-200 dark:border-slate-800">
          <h3 className="flex items-center gap-3 text-lg font-semibold text-slate-900 dark:text-white mb-4">
            <Layout className="h-5 w-5 text-orange-600" />
            Templates
          </h3>
          <p className="text-sm text-slate-600 dark:text-slate-400 mb-4">
            Page-level objects that place components into a layout and articulate the design's underlying content structure.
          </p>
          <ul className="text-sm space-y-1 text-slate-600 dark:text-slate-400">
            <li>• Page layouts</li>
            <li>• Auth page templates</li>
            <li>• Dashboard templates</li>
          </ul>
        </BaseCard>
      </div>

      <BaseCard className="border-slate-200 dark:border-slate-800">
        <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-6">Design Principles</h3>
        <div className="space-y-4">
          <div>
            <h4 className="font-medium text-slate-900 dark:text-white mb-2">Consistency</h4>
            <p className="text-sm text-slate-600 dark:text-slate-400">
              All components follow consistent patterns for spacing, colors, and interactions to ensure a cohesive user experience.
            </p>
          </div>
          <div>
            <h4 className="font-medium text-slate-900 dark:text-white mb-2">Accessibility</h4>
            <p className="text-sm text-slate-600 dark:text-slate-400">
              Components are built with accessibility in mind, supporting keyboard navigation and screen readers.
            </p>
          </div>
          <div>
            <h4 className="font-medium text-slate-900 dark:text-white mb-2">Dark Mode Support</h4>
            <p className="text-sm text-slate-600 dark:text-slate-400">
              Every component is designed to work seamlessly in both light and dark modes.
            </p>
          </div>
          <div>
            <h4 className="font-medium text-slate-900 dark:text-white mb-2">Server Components</h4>
            <p className="text-sm text-slate-600 dark:text-slate-400">
              Components are optimized for React Server Components to minimize client-side JavaScript.
            </p>
          </div>
        </div>
      </BaseCard>
    </div>
  );
}