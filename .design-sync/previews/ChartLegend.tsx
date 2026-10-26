import * as React from 'react'
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  type ChartConfig,
} from '@kloudlite/ui'
import { Area, AreaChart, Bar, BarChart, CartesianGrid, XAxis } from 'recharts'

const nodes = [
  { day: 'Mon', spot: 6, onDemand: 4 },
  { day: 'Tue', spot: 8, onDemand: 4 },
  { day: 'Wed', spot: 11, onDemand: 5 },
  { day: 'Thu', spot: 9, onDemand: 5 },
  { day: 'Fri', spot: 7, onDemand: 4 },
]

const config = {
  spot: { label: 'Spot nodes', color: 'var(--chart-1)' },
  onDemand: { label: 'On-demand nodes', color: 'var(--chart-2)' },
} satisfies ChartConfig

export const BottomLegend = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <BarChart data={nodes} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartLegend content={<ChartLegendContent />} />
      <Bar dataKey="spot" fill="var(--color-spot)" isAnimationActive={false} />
      <Bar dataKey="onDemand" fill="var(--color-onDemand)" isAnimationActive={false} />
    </BarChart>
  </ChartContainer>
)

export const TopLegend = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <AreaChart data={nodes} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartLegend verticalAlign="top" content={<ChartLegendContent />} />
      <Area
        dataKey="spot"
        stackId="a"
        stroke="var(--color-spot)"
        fill="var(--color-spot)"
        fillOpacity={0.25}
        isAnimationActive={false}
      />
      <Area
        dataKey="onDemand"
        stackId="a"
        stroke="var(--color-onDemand)"
        fill="var(--color-onDemand)"
        fillOpacity={0.25}
        isAnimationActive={false}
      />
    </AreaChart>
  </ChartContainer>
)

export const SingleSeries = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <BarChart data={nodes} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartLegend content={<ChartLegendContent />} />
      <Bar dataKey="spot" fill="var(--color-spot)" isAnimationActive={false} />
    </BarChart>
  </ChartContainer>
)
