import * as React from 'react'
import {
  Field,
  FieldDescription,
  FieldLabel,
  Input,
  Textarea,
} from '@kloudlite/ui'

export const InField = () => (
  <div className="w-96">
    <Field>
      <FieldLabel htmlFor="max-nodes">Maximum nodes</FieldLabel>
      <Input id="max-nodes" type="number" defaultValue={6} />
      <FieldDescription>
        The autoscaler will never grow this pool beyond this many nodes.
      </FieldDescription>
    </Field>
  </div>
)

export const WithLink = () => (
  <div className="w-96">
    <Field>
      <FieldLabel htmlFor="node-taints">Node taints</FieldLabel>
      <Textarea id="node-taints" defaultValue={'workload=build:NoSchedule'} />
      <FieldDescription>
        One taint per line. See the <a href="#taints-docs">scheduling guide</a> for accepted
        effects.
      </FieldDescription>
    </Field>
  </div>
)
