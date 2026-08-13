import * as React from 'react'
import { ChartContainer, type ChartConfig } from '@kloudlite/ui'
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Line,
  LineChart,
  XAxis,
  YAxis,
} from 'recharts'

const usage = [
  { hour: '09:00', cpu: 38, memory: 9.2 },
  { hour: '10:00', cpu: 52, memory: 11.4 },
  { hour: '11:00', cpu: 61, memory: 12.1 },
  { hour: '12:00', cpu: 47, memory: 10.8 },
  { hour: '13:00', cpu: 73, memory: 14.6 },
  { hour: '14:00', cpu: 66, memory: 13.2 },
]

const config = {
  cpu: { label: 'CPU %', color: 'var(--chart-1)' },
  memory: { label: 'Memory GiB', color: 'var(--chart-2)' },
} satisfies ChartConfig

export const LineSeries = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <LineChart data={usage} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="hour" tickLine={false} axisLine={false} />
      <YAxis tickLine={false} axisLine={false} width={32} />
      <Line dataKey="cpu" stroke="var(--color-cpu)" strokeWidth={2} dot={false} isAnimationActive={false} />
    </LineChart>
  </ChartContainer>
)

export const BarSeries = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <BarChart data={usage} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="hour" tickLine={false} axisLine={false} />
      <YAxis tickLine={false} axisLine={false} width={32} />
      <Bar dataKey="cpu" fill="var(--color-cpu)" isAnimationActive={false} />
      <Bar dataKey="memory" fill="var(--color-memory)" isAnimationActive={false} />
    </BarChart>
  </ChartContainer>
)

export const AreaSeries = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <AreaChart data={usage} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="hour" tickLine={false} axisLine={false} />
      <Area
        dataKey="memory"
        stroke="var(--color-memory)"
        fill="var(--color-memory)"
        fillOpacity={0.25}
        isAnimationActive={false}
      />
    </AreaChart>
  </ChartContainer>
)

export const CompactHeight = () => (
  <ChartContainer config={config} className="h-32 w-72">
    <LineChart data={usage} margin={{ top: 8, right: 8, bottom: 8, left: 8 }}>
      <Line dataKey="cpu" stroke="var(--color-cpu)" strokeWidth={2} dot={false} isAnimationActive={false} />
    </LineChart>
  </ChartContainer>
)
