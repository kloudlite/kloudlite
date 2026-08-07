import * as React from 'react'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@kloudlite/ui'
import { Bar, BarChart, CartesianGrid, Line, LineChart, XAxis, YAxis } from 'recharts'

const requests = [
  { hour: '09:00', ingress: 1240, egress: 810 },
  { hour: '10:00', ingress: 1680, egress: 940 },
  { hour: '11:00', ingress: 2110, egress: 1220 },
  { hour: '12:00', ingress: 1890, egress: 1105 },
  { hour: '13:00', ingress: 2470, egress: 1380 },
]

const config = {
  ingress: { label: 'Ingress', color: 'var(--chart-1)' },
  egress: { label: 'Egress', color: 'var(--chart-2)' },
} satisfies ChartConfig

// defaultIndex pins the tooltip open so the static render shows it.
export const OnLineChart = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <LineChart data={requests} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="hour" tickLine={false} axisLine={false} />
      <YAxis tickLine={false} axisLine={false} width={40} />
      <ChartTooltip defaultIndex={2} content={<ChartTooltipContent />} />
      <Line dataKey="ingress" stroke="var(--color-ingress)" strokeWidth={2} dot={false} isAnimationActive={false} />
    </LineChart>
  </ChartContainer>
)

export const MultiSeries = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <BarChart data={requests} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="hour" tickLine={false} axisLine={false} />
      <ChartTooltip defaultIndex={1} content={<ChartTooltipContent />} />
      <Bar dataKey="ingress" fill="var(--color-ingress)" isAnimationActive={false} />
      <Bar dataKey="egress" fill="var(--color-egress)" isAnimationActive={false} />
    </BarChart>
  </ChartContainer>
)

export const WithoutCursor = () => (
  <ChartContainer config={config} className="h-64 w-96">
    <LineChart data={requests} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
      <CartesianGrid vertical={false} />
      <XAxis dataKey="hour" tickLine={false} axisLine={false} />
      <ChartTooltip
        cursor={false}
        defaultIndex={4}
        content={<ChartTooltipContent hideLabel />}
      />
      <Line dataKey="egress" stroke="var(--color-egress)" strokeWidth={2} dot={false} isAnimationActive={false} />
    </LineChart>
  </ChartContainer>
)
