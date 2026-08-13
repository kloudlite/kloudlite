import * as React from 'react'
import { Calendar } from '@kloudlite/ui'

// CalendarDayButton is the day cell renderer used by Calendar; it needs the
// DayPicker context to receive `day`/`modifiers`, so each preview renders the
// full Calendar with the relevant day state applied.
const march = new Date(2026, 2, 1)

export const SelectedDay = () => (
  <Calendar
    mode="single"
    defaultMonth={march}
    selected={new Date(2026, 2, 12)}
    onSelect={() => {}}
  />
)

export const RangeStartMiddleEnd = () => (
  <Calendar
    mode="range"
    defaultMonth={march}
    selected={{ from: new Date(2026, 2, 10), to: new Date(2026, 2, 18) }}
    onSelect={() => {}}
  />
)

export const DisabledDays = () => (
  <Calendar
    mode="single"
    defaultMonth={march}
    selected={new Date(2026, 2, 5)}
    disabled={{ before: new Date(2026, 2, 3) }}
    onSelect={() => {}}
  />
)

export const MultipleSelected = () => (
  <Calendar
    mode="multiple"
    defaultMonth={march}
    selected={[new Date(2026, 2, 3), new Date(2026, 2, 11), new Date(2026, 2, 24)]}
    onSelect={() => {}}
  />
)
