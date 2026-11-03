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

export const OpenFlatList = () => (
  <div className="flex h-64 w-80 flex-col">
    <Select open defaultValue="ap-south-1">
      <SelectTrigger>
        <SelectValue placeholder="Select a region" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="ap-south-1">ap-south-1 &middot; Mumbai</SelectItem>
        <SelectItem value="eu-west-1">eu-west-1 &middot; Ireland</SelectItem>
        <SelectItem value="us-east-1">us-east-1 &middot; N. Virginia</SelectItem>
      </SelectContent>
    </Select>
  </div>
)

export const OpenGrouped = () => (
  <div className="flex h-72 w-80 flex-col">
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

export const Closed = () => (
  <div className="w-80">
    <Select defaultValue="staging">
      <SelectTrigger>
        <SelectValue placeholder="Select an environment" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="dev">development</SelectItem>
        <SelectItem value="staging">staging</SelectItem>
      </SelectContent>
    </Select>
  </div>
)
