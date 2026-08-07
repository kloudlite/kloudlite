import * as React from 'react'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'

export const BetweenGroups = () => (
  <div className="flex h-80 w-80 flex-col">
    <Select open defaultValue="medium">
      <SelectTrigger>
        <SelectValue placeholder="Select a cluster size" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>Shared</SelectLabel>
          <SelectItem value="small">Small &middot; 2 vCPU</SelectItem>
          <SelectItem value="medium">Medium &middot; 4 vCPU</SelectItem>
        </SelectGroup>
        <SelectSeparator />
        <SelectGroup>
          <SelectLabel>Dedicated</SelectLabel>
          <SelectItem value="large">Large &middot; 8 vCPU</SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  </div>
)

export const BeforeTrailingAction = () => (
  <div className="flex h-72 w-80 flex-col">
    <Select open defaultValue="kl-prod">
      <SelectTrigger>
        <SelectValue placeholder="Select a cluster" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="kl-prod">kl-prod &middot; ap-south-1</SelectItem>
        <SelectItem value="kl-staging">kl-staging &middot; eu-west-1</SelectItem>
        <SelectSeparator />
        <SelectItem value="new">Create a new cluster</SelectItem>
      </SelectContent>
    </Select>
  </div>
)
