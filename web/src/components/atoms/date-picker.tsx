"use client"

import * as React from "react"
import { format } from "date-fns"
import { Calendar as CalendarIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { Calendar } from "@/components/ui/calendar"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/molecules/popover"

interface DatePickerProps {
  date?: Date
  onDateChange?: (date: Date | undefined) => void
  placeholder?: string
  disabled?: boolean
  className?: string
  error?: boolean
}

export function DatePicker({
  date,
  onDateChange,
  placeholder = "Pick a date",
  disabled = false,
  className,
  error,
}: DatePickerProps) {
  const [open, setOpen] = React.useState(false)

  const baseStyles = [
    // Base styles
    "flex h-10 w-full items-center justify-between rounded-md",
    "px-3 py-2",
    "text-sm",
    
    // Background
    "bg-background",
    
    // Border
    "border",
    
    // Typography
    "text-foreground",
    
    // Focus styles
    "outline-none",
    
    // Transitions
    "transition-all duration-200",
    
    // Disabled
    "disabled:cursor-not-allowed disabled:opacity-70 disabled:bg-muted/30",
    
    // Interactive
    "cursor-pointer",
  ];

  const normalStyles = [
    "border-form-border",
    "hover:border-form-border-hover",
    "focus:border-form-border-focus",
    "focus:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_rgb(var(--color-brand-500))]"
  ];

  const errorStyles = [
    "border-error",
    "hover:border-error",
    "focus:border-error",
    "focus:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_var(--color-error)]"
  ];

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          className={cn(
            baseStyles,
            error ? errorStyles : normalStyles,
            className
          )}
          disabled={disabled}
          aria-invalid={error}
        >
          <span className={cn(!date && "text-muted-foreground")}>
            {date ? format(date, "PPP") : placeholder}
          </span>
          <CalendarIcon className="h-4 w-4 opacity-50" />
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="single"
          selected={date}
          onSelect={(selectedDate) => {
            onDateChange?.(selectedDate)
            setOpen(false)
          }}
          initialFocus
        />
      </PopoverContent>
    </Popover>
  )
}

interface DateRangePickerProps {
  dateRange?: { from?: Date; to?: Date }
  onDateRangeChange?: (range: { from?: Date; to?: Date } | undefined) => void
  placeholder?: string
  disabled?: boolean
  className?: string
  error?: boolean
}

export function DateRangePicker({
  dateRange,
  onDateRangeChange,
  placeholder = "Pick a date range",
  disabled = false,
  className,
  error,
}: DateRangePickerProps) {
  const [open, setOpen] = React.useState(false)

  const baseStyles = [
    // Base styles
    "flex h-10 w-full items-center justify-between rounded-md",
    "px-3 py-2",
    "text-sm",
    
    // Background
    "bg-background",
    
    // Border
    "border",
    
    // Typography
    "text-foreground",
    
    // Focus styles
    "outline-none",
    
    // Transitions
    "transition-all duration-200",
    
    // Disabled
    "disabled:cursor-not-allowed disabled:opacity-70 disabled:bg-muted/30",
    
    // Interactive
    "cursor-pointer",
  ];

  const normalStyles = [
    "border-form-border",
    "hover:border-form-border-hover",
    "focus:border-form-border-focus",
    "focus:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_rgb(var(--color-brand-500))]"
  ];

  const errorStyles = [
    "border-error",
    "hover:border-error",
    "focus:border-error",
    "focus:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_var(--color-error)]"
  ];

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          className={cn(
            baseStyles,
            error ? errorStyles : normalStyles,
            className
          )}
          disabled={disabled}
          aria-invalid={error}
        >
          <span className={cn(!dateRange?.from && "text-muted-foreground")}>
            {dateRange?.from ? (
              dateRange.to ? (
                <>
                  {format(dateRange.from, "LLL dd, y")} -{" "}
                  {format(dateRange.to, "LLL dd, y")}
                </>
              ) : (
                format(dateRange.from, "LLL dd, y")
              )
            ) : (
              placeholder
            )}
          </span>
          <CalendarIcon className="h-4 w-4 opacity-50" />
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          initialFocus
          mode="range"
          defaultMonth={dateRange?.from}
          selected={dateRange}
          onSelect={onDateRangeChange}
          numberOfMonths={2}
        />
      </PopoverContent>
    </Popover>
  )
}

interface DateTimePickerProps {
  date?: Date
  onDateChange?: (date: Date | undefined) => void
  placeholder?: string
  disabled?: boolean
  className?: string
  showTime?: boolean
  error?: boolean
}

export function DateTimePicker({
  date,
  onDateChange,
  placeholder = "Pick a date and time",
  disabled = false,
  className,
  showTime = true,
  error,
}: DateTimePickerProps) {
  const [open, setOpen] = React.useState(false)
  const [selectedDate, setSelectedDate] = React.useState<Date | undefined>(date)
  const [timeValue, setTimeValue] = React.useState<string>(
    date ? format(date, "HH:mm") : "00:00"
  )

  const handleDateSelect = (newDate: Date | undefined) => {
    if (newDate) {
      // Preserve the time when selecting a new date
      const [hours, minutes] = timeValue.split(":").map(Number)
      newDate.setHours(hours, minutes, 0, 0)
      setSelectedDate(newDate)
      onDateChange?.(newDate)
    } else {
      setSelectedDate(undefined)
      onDateChange?.(undefined)
    }
  }

  const handleTimeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newTime = event.target.value
    setTimeValue(newTime)
    
    if (selectedDate) {
      const [hours, minutes] = newTime.split(":").map(Number)
      const newDateTime = new Date(selectedDate)
      newDateTime.setHours(hours, minutes, 0, 0)
      setSelectedDate(newDateTime)
      onDateChange?.(newDateTime)
    }
  }

  const baseStyles = [
    // Base styles
    "flex h-10 w-full items-center justify-between rounded-md",
    "px-3 py-2",
    "text-sm",
    
    // Background
    "bg-background",
    
    // Border
    "border",
    
    // Typography
    "text-foreground",
    
    // Focus styles
    "outline-none",
    
    // Transitions
    "transition-all duration-200",
    
    // Disabled
    "disabled:cursor-not-allowed disabled:opacity-70 disabled:bg-muted/30",
    
    // Interactive
    "cursor-pointer",
  ];

  const normalStyles = [
    "border-form-border",
    "hover:border-form-border-hover",
    "focus:border-form-border-focus",
    "focus:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_rgb(var(--color-brand-500))]"
  ];

  const errorStyles = [
    "border-error",
    "hover:border-error",
    "focus:border-error",
    "focus:[box-shadow:0_0_0_2px_var(--background),0_0_0_4px_var(--color-error)]"
  ];

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          className={cn(
            baseStyles,
            error ? errorStyles : normalStyles,
            className
          )}
          disabled={disabled}
          aria-invalid={error}
        >
          <span className={cn(!selectedDate && "text-muted-foreground")}>
            {selectedDate ? (
              showTime ? (
                format(selectedDate, "PPP 'at' HH:mm")
              ) : (
                format(selectedDate, "PPP")
              )
            ) : (
              placeholder
            )}
          </span>
          <CalendarIcon className="h-4 w-4 opacity-50" />
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <div className="p-3 space-y-3">
          <Calendar
            mode="single"
            selected={selectedDate}
            onSelect={handleDateSelect}
            initialFocus
          />
          {showTime && (
            <div className="flex items-center gap-2 px-3 py-2 border-t">
              <span className="text-sm text-muted-foreground">Time:</span>
              <input
                type="time"
                value={timeValue}
                onChange={handleTimeChange}
                className="text-sm border rounded px-2 py-1 bg-background focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
}