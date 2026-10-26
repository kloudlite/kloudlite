import * as React from 'react'
import { ScrollArea } from '@kloudlite/ui'

const events = [
  'FailedScheduling — 0/3 nodes available',
  'Scheduled — assigned to node ip-10-0-2-14',
  'Pulling — ghcr.io/kloudlite/api:v1.2.9',
  'Pulled — image pulled in 4.2s',
  'Created — container api',
  'Started — container api',
  'Unhealthy — readiness probe failed',
  'Healthy — readiness probe succeeded',
  'Killing — old revision draining',
  'ScalingReplicaSet — scaled up to 3',
  'SuccessfulCreate — pod api-gateway-7d9f4c',
  'BackOff — restarting failed container',
]

export const VerticalScrollBar = () => (
  <ScrollArea type="always" className="h-64 w-80 border">
    <div className="p-3 text-sm">
      {events.map((e, i) => (
        <p key={i} className="border-b py-2">
          {e}
        </p>
      ))}
    </div>
  </ScrollArea>
)

export const DenseListScrollBar = () => (
  <ScrollArea type="always" className="h-72 w-96 border bg-muted">
    <div className="p-3 font-mono text-xs leading-relaxed">
      {events.concat(events).map((e, i) => (
        <p key={i}>{e}</p>
      ))}
    </div>
  </ScrollArea>
)
