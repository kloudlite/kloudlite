import * as React from 'react'
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  type ChartConfig,
} from '@kloudlite/ui'
import { Bar, BarChart, CartesianGrid, Line, LineChart, XAxis } from 'recharts'

const traffic = [
  { day: 'Mon', tunnel: 420, intercept: 180, direct: 96 },
  { day: 'Tue', tunnel: 512, intercept: 204, direct: 88 },
  { day: 'Wed', tunnel: 604, intercept: 261, direct: 112 },
  { day: 'Thu', tunnel: 548, intercept: 232, direct: 104 },
  { day: 'Fri', tunnel: 470, intercept: 199, direct: 91 },
]

const config = {
  tunnel: { label: 'Tunnel MB', color: 'var(--chart-1)' },
  intercept: { label: 'Intercept MB', color: 'var(--chart-2)' },
  direct: { label: 'Direct MB', color: 'var(--chart-3)' },
} satisfies ChartConfig

export const ThreeSeries = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <BarChart data={traffic} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartLegend content={<ChartLegendContent />} />
      <Bar dataKey="tunnel" fill="var(--color-tunnel)" isAnimationActive={false} />
      <Bar dataKey="intercept" fill="var(--color-intercept)" isAnimationActive={false} />
      <Bar dataKey="direct" fill="var(--color-direct)" isAnimationActive={false} />
    </BarChart>
  </ChartContainer>
)

export const AlignedTop = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <LineChart data={traffic} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartLegend verticalAlign="top" content={<ChartLegendContent />} />
      <Line dataKey="tunnel" stroke="var(--color-tunnel)" strokeWidth={2} dot={false} isAnimationActive={false} />
      <Line
        dataKey="intercept"
        stroke="var(--color-intercept)"
        strokeWidth={2}
        dot={false} isAnimationActive={false}
      />
    </LineChart>
  </ChartContainer>
)

export const WithoutSwatches = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <BarChart data={traffic} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="day" tickLine={false} axisLine={false} />
      <ChartLegend content={<ChartLegendContent hideIcon />} />
      <Bar dataKey="tunnel" fill="var(--color-tunnel)" isAnimationActive={false} />
      <Bar dataKey="intercept" fill="var(--color-intercept)" isAnimationActive={false} />
    </BarChart>
  </ChartContainer>
)
