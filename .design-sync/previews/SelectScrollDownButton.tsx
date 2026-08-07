import * as React from 'react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'

const regions = [
  'ap-south-1 · Mumbai',
  'ap-southeast-1 · Singapore',
  'ap-northeast-1 · Tokyo',
  'eu-west-1 · Ireland',
  'eu-central-1 · Frankfurt',
  'us-east-1 · N. Virginia',
  'us-west-2 · Oregon',
  'sa-east-1 · São Paulo',
  'ca-central-1 · Montreal',
  'me-south-1 · Bahrain',
]

export const OverflowingList = () => (
  <div className="flex h-64 w-80 flex-col">
    <Select open defaultValue={regions[0]}>
      <SelectTrigger>
        <SelectValue placeholder="Select a region" />
      </SelectTrigger>
      <SelectContent className="max-h-64">
        {regions.map((r) => (
          <SelectItem key={r} value={r}>
            {r}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  </div>
)

const pools = [
  'ondemand-2x · 2 vCPU',
  'ondemand-4x · 4 vCPU',
  'ondemand-8x · 8 vCPU',
  'ondemand-16x · 16 vCPU',
  'spot-2x · 2 vCPU',
  'spot-4x · 4 vCPU',
  'spot-8x · 8 vCPU',
  'gpu-a10 · 1 GPU',
  'gpu-a100 · 1 GPU',
  'mem-8x · 64 GiB',
]

export const CompactList = () => (
  <div className="flex h-64 w-80 flex-col">
    <Select open defaultValue={pools[0]}>
      <SelectTrigger>
        <SelectValue placeholder="Select a node pool" />
      </SelectTrigger>
      <SelectContent className="max-h-64">
        {pools.map((p) => (
          <SelectItem key={p} value={p}>
            {p}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  </div>
)
