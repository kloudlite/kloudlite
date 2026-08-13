import * as React from 'react'
import {
  Label,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'

export const Placeholder = () => (
  <div className="w-80 space-y-2">
    <Label htmlFor="sv-env">Environment</Label>
    <Select>
      <SelectTrigger id="sv-env">
        <SelectValue placeholder="Select an environment" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="dev">development</SelectItem>
        <SelectItem value="staging">staging</SelectItem>
        <SelectItem value="prod">production</SelectItem>
      </SelectContent>
    </Select>
  </div>
)

export const SelectedValue = () => (
  <div className="w-80 space-y-2">
    <Label htmlFor="sv-region">Region</Label>
    <Select defaultValue="ap-south-1">
      <SelectTrigger id="sv-region">
        <SelectValue placeholder="Select a region" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="ap-south-1">ap-south-1 &middot; Mumbai</SelectItem>
        <SelectItem value="eu-west-1">eu-west-1 &middot; Ireland</SelectItem>
      </SelectContent>
    </Select>
  </div>
)

export const LongValueTruncated = () => (
  <div className="w-64">
    <Select defaultValue="img">
      <SelectTrigger>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="img">
          ghcr.io/kloudlite/workspace-base:v1.0.8-amd64
        </SelectItem>
      </SelectContent>
    </Select>
  </div>
)
