import * as React from 'react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'

const tags = [
  'v1.0.0',
  'v1.0.1',
  'v1.0.2',
  'v1.0.3',
  'v1.0.4',
  'v1.0.5',
  'v1.0.6',
  'v1.0.7',
  'v1.0.8',
  'v1.0.9',
]

export const ScrolledIntoList = () => (
  <div className="flex h-64 w-80 flex-col justify-end">
    <Select open defaultValue="v1.0.9">
      <SelectTrigger>
        <SelectValue placeholder="Select an image tag" />
      </SelectTrigger>
      <SelectContent position="item-aligned" className="max-h-64">
        {tags.map((t) => (
          <SelectItem key={t} value={t}>
            workspace-base:{t}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  </div>
)

const versions = [
  '1.26',
  '1.27',
  '1.28',
  '1.29',
  '1.30',
  '1.31',
  '1.32',
  '1.33',
  '1.34',
  '1.35',
]

export const LastOptionSelected = () => (
  <div className="flex h-64 w-80 flex-col justify-end">
    <Select open defaultValue="1.35">
      <SelectTrigger>
        <SelectValue placeholder="Select a Kubernetes version" />
      </SelectTrigger>
      <SelectContent position="item-aligned" className="max-h-64">
        {versions.map((v) => (
          <SelectItem key={v} value={v}>
            Kubernetes {v}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  </div>
)
