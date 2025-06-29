"use client";

import { useState } from "react";
import { ComponentShowcase } from "../../_components/component-showcase";
import { Text, Heading } from "@/components/atoms";
import { DatePicker, DateRangePicker, DateTimePicker } from "@/components/atoms";

export default function DatePickerPage() {
  const [selectedDate, setSelectedDate] = useState<Date | undefined>();
  const [selectedDateRange, setSelectedDateRange] = useState<{ from?: Date; to?: Date } | undefined>();
  const [selectedDateTime, setSelectedDateTime] = useState<Date | undefined>();

  return (
    <div className="space-y-8">
      <div>
        <Heading level={2} className="mb-4">Date Picker</Heading>
        <Text color="secondary">
          Date picker components built with react-day-picker and date-fns for consistent date selection.
        </Text>
      </div>

      {/* Single Date Picker */}
      <ComponentShowcase
        title="Date Picker"
        description="Single date selection with calendar popup"
      >
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-3">
              <Text weight="medium">Default Date Picker</Text>
              <DatePicker
                date={selectedDate}
                onDateChange={setSelectedDate}
                placeholder="Select a date"
              />
              {selectedDate && (
                <Text size="sm" color="muted">
                  Selected: {selectedDate.toLocaleDateString()}
                </Text>
              )}
            </div>

            <div className="space-y-3">
              <Text weight="medium">Pre-selected Date</Text>
              <DatePicker
                date={new Date()}
                onDateChange={() => {}}
                placeholder="Today's date"
              />
            </div>

            <div className="space-y-3">
              <Text weight="medium">Disabled State</Text>
              <DatePicker
                disabled
                placeholder="Disabled date picker"
              />
            </div>

            <div className="space-y-3">
              <Text weight="medium">Custom Placeholder</Text>
              <DatePicker
                placeholder="Choose your birthday"
              />
            </div>
          </div>
        </div>
      </ComponentShowcase>

      {/* Date Range Picker */}
      <ComponentShowcase
        title="Date Range Picker"
        description="Select a range of dates with dual calendar view"
      >
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-3">
              <Text weight="medium">Date Range Selection</Text>
              <DateRangePicker
                dateRange={selectedDateRange}
                onDateRangeChange={setSelectedDateRange}
                placeholder="Select date range"
              />
              {selectedDateRange?.from && (
                <Text size="sm" color="muted">
                  From: {selectedDateRange.from.toLocaleDateString()}
                  {selectedDateRange.to && ` - To: ${selectedDateRange.to.toLocaleDateString()}`}
                </Text>
              )}
            </div>

            <div className="space-y-3">
              <Text weight="medium">Vacation Period</Text>
              <DateRangePicker
                placeholder="Select vacation dates"
              />
            </div>

            <div className="space-y-3">
              <Text weight="medium">Disabled Range Picker</Text>
              <DateRangePicker
                disabled
                placeholder="Disabled range picker"
              />
            </div>

            <div className="space-y-3">
              <Text weight="medium">Project Timeline</Text>
              <DateRangePicker
                placeholder="Select project duration"
              />
            </div>
          </div>
        </div>
      </ComponentShowcase>

      {/* Date Time Picker */}
      <ComponentShowcase
        title="Date Time Picker"
        description="Date and time selection with time input"
      >
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-3">
              <Text weight="medium">Date & Time Picker</Text>
              <DateTimePicker
                date={selectedDateTime}
                onDateChange={setSelectedDateTime}
                placeholder="Select date and time"
              />
              {selectedDateTime && (
                <Text size="sm" color="muted">
                  Selected: {selectedDateTime.toLocaleString()}
                </Text>
              )}
            </div>

            <div className="space-y-3">
              <Text weight="medium">Date Only (No Time)</Text>
              <DateTimePicker
                showTime={false}
                placeholder="Select date only"
              />
            </div>

            <div className="space-y-3">
              <Text weight="medium">Meeting Scheduler</Text>
              <DateTimePicker
                placeholder="Schedule meeting"
              />
            </div>

            <div className="space-y-3">
              <Text weight="medium">Deadline Picker</Text>
              <DateTimePicker
                placeholder="Set deadline"
              />
            </div>
          </div>
        </div>
      </ComponentShowcase>

      {/* Sizes and Variants */}
      <ComponentShowcase
        title="Custom Styling"
        description="Date pickers with custom styling and sizes"
      >
        <div className="space-y-6">
          <div>
            <Text weight="medium" className="mb-4">Different Widths</Text>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <DatePicker
                placeholder="Small width"
                className="max-w-[200px]"
              />
              <DatePicker
                placeholder="Medium width"
                className="max-w-[300px]"
              />
              <DatePicker
                placeholder="Full width"
              />
            </div>
          </div>

          <div>
            <Text weight="medium" className="mb-4">Form Integration</Text>
            <div className="max-w-md space-y-4">
              <div>
                <label className="text-sm font-medium mb-2 block">Birth Date</label>
                <DatePicker placeholder="Enter your birth date" />
              </div>
              <div>
                <label className="text-sm font-medium mb-2 block">Event Date Range</label>
                <DateRangePicker placeholder="Select event duration" />
              </div>
              <div>
                <label className="text-sm font-medium mb-2 block">Appointment Time</label>
                <DateTimePicker placeholder="Schedule appointment" />
              </div>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      {/* Usage Examples */}
      <ComponentShowcase
        title="Usage Examples"
        description="Common use cases and implementation patterns"
      >
        <div className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="space-y-4">
              <Text weight="medium">Event Planning</Text>
              <div className="space-y-3">
                <DatePicker placeholder="Event date" />
                <DateTimePicker placeholder="Event start time" />
                <DateTimePicker placeholder="Event end time" />
              </div>
            </div>

            <div className="space-y-4">
              <Text weight="medium">Booking System</Text>
              <div className="space-y-3">
                <DateRangePicker placeholder="Check-in to check-out" />
                <DateTimePicker placeholder="Arrival time" />
                <DatePicker placeholder="Departure date" />
              </div>
            </div>

            <div className="space-y-4">
              <Text weight="medium">Report Generation</Text>
              <div className="space-y-3">
                <DateRangePicker placeholder="Report period" />
                <DatePicker placeholder="Generate date" />
              </div>
            </div>

            <div className="space-y-4">
              <Text weight="medium">Task Management</Text>
              <div className="space-y-3">
                <DatePicker placeholder="Due date" />
                <DateTimePicker placeholder="Reminder time" />
                <DateRangePicker placeholder="Project timeline" />
              </div>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      {/* Code Examples */}
      <ComponentShowcase
        title="Implementation"
        description="How to use the date picker components"
      >
        <div className="space-y-6">
          <div className="bg-muted p-4 rounded-lg">
            <Text weight="medium" className="mb-3">Basic Usage</Text>
            <pre className="text-sm overflow-x-auto">
              <code>{`import { DatePicker, DateRangePicker, DateTimePicker } from '@/components/atoms';

// Single date picker
<DatePicker
  date={selectedDate}
  onDateChange={setSelectedDate}
  placeholder="Select a date"
/>

// Date range picker
<DateRangePicker
  dateRange={selectedRange}
  onDateRangeChange={setSelectedRange}
  placeholder="Select date range"
/>

// Date time picker
<DateTimePicker
  date={selectedDateTime}
  onDateChange={setSelectedDateTime}
  placeholder="Select date and time"
  showTime={true}
/>`}</code>
            </pre>
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}