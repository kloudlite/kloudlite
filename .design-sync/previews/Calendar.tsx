import * as React from 'react'
import { Calendar } from '@kloudlite/ui'

const march = new Date(2026, 2, 1)

export const SingleSelection = () => (
  <Calendar
    mode="single"
    defaultMonth={march}
    selected={new Date(2026, 2, 12)}
    onSelect={() => {}}
  />
)

export const RangeSelection = () => (
  <Calendar
    mode="range"
    defaultMonth={march}
    selected={{ from: new Date(2026, 2, 9), to: new Date(2026, 2, 16) }}
    onSelect={() => {}}
  />
)

export const WithDisabledDays = () => (
  <Calendar
    mode="single"
    defaultMonth={march}
    selected={new Date(2026, 2, 18)}
    disabled={{ after: new Date(2026, 2, 20) }}
    onSelect={() => {}}
  />
)

export const DropdownCaption = () => (
  <Calendar
    mode="single"
    captionLayout="dropdown"
    defaultMonth={march}
    startMonth={new Date(2025, 0, 1)}
    endMonth={new Date(2026, 11, 1)}
    selected={new Date(2026, 2, 4)}
    onSelect={() => {}}
  />
)
