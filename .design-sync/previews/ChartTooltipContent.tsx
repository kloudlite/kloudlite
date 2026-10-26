import * as React from 'react'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@kloudlite/ui'
import { Bar, BarChart, CartesianGrid, Line, LineChart, XAxis } from 'recharts'

const builds = [
  { day: 'Mon', api: 164, console: 212 },
  { day: 'Tue', api: 149, console: 198 },
  { day: 'Wed', api: 181, console: 241 },
  { day: 'Thu', api: 172, console: 226 },
  { day: 'Fri', api: 158, console: 205 },
]

const config = {
  api: { label: 'api (s)', color: 'var(--chart-1)' },
  console: { label: 'console (s)', color: 'var(--chart-2)' },
} satisfies ChartConfig

// defaultIndex pins the tooltip open so the static render shows it.
export const DotIndicator = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <BarChart data={builds} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartTooltip defaultIndex={2} content={<ChartTooltipContent indicator="dot" />} />
      <Bar dataKey="api" fill="var(--color-api)" isAnimationActive={false} />
      <Bar dataKey="console" fill="var(--color-console)" isAnimationActive={false} />
    </BarChart>
  </ChartContainer>
)

export const LineIndicator = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <LineChart data={builds} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartTooltip defaultIndex={1} content={<ChartTooltipContent indicator="line" />} />
      <Line dataKey="api" stroke="var(--color-api)" strokeWidth={2} dot={false} isAnimationActive={false} />
    </LineChart>
  </ChartContainer>
)

export const DashedIndicator = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <LineChart data={builds} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartTooltip
        defaultIndex={3}
        content={<ChartTooltipContent indicator="dashed" />}
      />
      <Line dataKey="console" stroke="var(--color-console)" strokeWidth={2} dot={false} isAnimationActive={false} />
    </LineChart>
  </ChartContainer>
)

export const HiddenLabel = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <BarChart data={builds} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartTooltip defaultIndex={4} content={<ChartTooltipContent hideLabel />} />
      <Bar dataKey="api" fill="var(--color-api)" isAnimationActive={false} />
    </BarChart>
  </ChartContainer>
)
