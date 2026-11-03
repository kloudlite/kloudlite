import * as React from 'react'
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  type ChartConfig,
} from '@kloudlite/ui'
import { Bar, BarChart, CartesianGrid, Line, LineChart, XAxis } from 'recharts'

// ChartStyle is rendered internally by ChartContainer: it emits the
// --color-<key> custom properties that the marks below reference.
const pods = [
  { day: 'Mon', running: 42, pending: 5, failed: 2 },
  { day: 'Tue', running: 48, pending: 3, failed: 1 },
  { day: 'Wed', running: 55, pending: 6, failed: 3 },
  { day: 'Thu', running: 51, pending: 4, failed: 1 },
  { day: 'Fri', running: 46, pending: 2, failed: 0 },
]

const tokenConfig = {
  running: { label: 'Running pods', color: 'var(--chart-1)' },
  pending: { label: 'Pending pods', color: 'var(--chart-4)' },
  failed: { label: 'Failed pods', color: 'var(--chart-3)' },
} satisfies ChartConfig

const themedConfig = {
  running: {
    label: 'Running pods',
    theme: { light: 'var(--chart-2)', dark: 'var(--chart-5)' },
  },
} satisfies ChartConfig

export const ColorsFromConfig = () => (
  <ChartContainer config={tokenConfig} className="h-64 w-96">
    <BarChart data={pods} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartLegend content={<ChartLegendContent />} />
      <Bar dataKey="running" fill="var(--color-running)" isAnimationActive={false} />
      <Bar dataKey="pending" fill="var(--color-pending)" isAnimationActive={false} />
      <Bar dataKey="failed" fill="var(--color-failed)" isAnimationActive={false} />
    </BarChart>
  </ChartContainer>
)

export const ThemedColor = () => (
  <ChartContainer config={themedConfig} className="h-64 w-96">
    <LineChart data={pods} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartLegend content={<ChartLegendContent />} />
      <Line dataKey="running" stroke="var(--color-running)" strokeWidth={2} dot={false} isAnimationActive={false} />
    </LineChart>
  </ChartContainer>
)

export const ScopedPerChart = () => (
  <div className="flex gap-4">
    <ChartContainer config={tokenConfig} className="h-48 w-72">
      <BarChart data={pods} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
        <XAxis dataKey="day" tickLine={false} axisLine={false} />
        <Bar dataKey="running" fill="var(--color-running)" isAnimationActive={false} />
      </BarChart>
    </ChartContainer>
    <ChartContainer config={themedConfig} className="h-48 w-72">
      <BarChart data={pods} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
        <XAxis dataKey="day" tickLine={false} axisLine={false} />
        <Bar dataKey="running" fill="var(--color-running)" isAnimationActive={false} />
      </BarChart>
    </ChartContainer>
  </div>
)
