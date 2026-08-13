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

export const SingleGroup = () => (
  <div className="flex h-64 w-80 flex-col">
    <Select open defaultValue="dev">
      <SelectTrigger>
        <SelectValue placeholder="Select an environment" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>Team environments</SelectLabel>
          <SelectItem value="dev">development</SelectItem>
          <SelectItem value="staging">staging</SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  </div>
)

export const TwoGroupsSeparated = () => (
  <div className="flex h-80 w-80 flex-col">
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
        <SelectSeparator />
        <SelectGroup>
          <SelectLabel>Europe</SelectLabel>
          <SelectItem value="eu-west-1">eu-west-1 &middot; Ireland</SelectItem>
          <SelectItem value="eu-central-1">
            eu-central-1 &middot; Frankfurt
          </SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  </div>
)
