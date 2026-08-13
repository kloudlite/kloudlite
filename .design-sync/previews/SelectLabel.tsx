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

export const SingleSectionLabel = () => (
  <div className="flex h-64 w-80 flex-col">
    <Select open defaultValue="ap-south-1">
      <SelectTrigger>
        <SelectValue placeholder="Select a region" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>Asia Pacific</SelectLabel>
          <SelectItem value="ap-south-1">ap-south-1 &middot; Mumbai</SelectItem>
          <SelectItem value="ap-southeast-1">
            ap-southeast-1 &middot; Singapore
          </SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  </div>
)

export const MultipleSectionLabels = () => (
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
          <SelectItem value="xlarge">XLarge &middot; 16 vCPU</SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  </div>
)
