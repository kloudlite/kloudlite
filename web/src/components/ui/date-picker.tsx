"use client"

import * as React from "react"
import { format } from "date-fns"
import { Calendar as CalendarIcon, X } from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

interface DatePickerProps {
  date?: Date
  onSelect?: (date: Date | undefined) => void
  placeholder?: string
  disabled?: boolean
  buttonClassName?: string
  calendarClassName?: string
  formatStr?: string
  disabledDates?: any
  fromDate?: Date
  toDate?: Date
}

export function DatePicker({
  date,
  onSelect,
  placeholder = "Pick a date",
  disabled = false,
  buttonClassName,
  calendarClassName,
  formatStr = "PPP",
  disabledDates,
  fromDate,
  toDate,
}: DatePickerProps) {
  const [open, setOpen] = React.useState(false)
  const [isCalendarHovered, setIsCalendarHovered] = React.useState(false)
  const [displayMonth, setDisplayMonth] = React.useState<Date>(date || new Date())

  const handleSelect = (selectedDate: Date | undefined) => {
    onSelect?.(selectedDate)
    // Add a small delay before closing to show the selection animation
    setTimeout(() => {
      setOpen(false)
    }, 150)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          disabled={disabled}
          className={cn(
            "w-[280px] justify-start text-left font-normal",
            "",
            "transition-all duration-200",
            "active:scale-100", // Override the default button scale
            !date && "text-muted-foreground",
            disabled && "opacity-50 cursor-not-allowed",
            open && "ring-2 ring-ring ring-offset-2 ring-offset-background",
            buttonClassName
          )}
        >
          <CalendarIcon 
            className={cn(
              "mr-2 h-4 w-4",
              "transition-transform duration-200",
              open && "rotate-12"
            )} 
          />
          {date ? format(date, formatStr) : <span>{placeholder}</span>}
        </Button>
      </PopoverTrigger>
      <PopoverContent 
        className={cn(
          "w-auto p-0",
          "animate-in fade-in-0",
          "data-[side=bottom]:slide-in-from-top-2",
          "data-[side=left]:slide-in-from-right-2",
          "data-[side=right]:slide-in-from-left-2",
          "data-[side=top]:slide-in-from-bottom-2"
        )}
        align="start"
        onMouseEnter={() => setIsCalendarHovered(true)}
        onMouseLeave={() => setIsCalendarHovered(false)}
      >
        {/* Year/Month Selection - Always visible */}
        <div className="flex items-center justify-center gap-2 p-3 pb-0">
          <Select 
            value={displayMonth.getMonth().toString()} 
            onValueChange={(value) => {
              const newDate = new Date(displayMonth)
              newDate.setMonth(parseInt(value))
              setDisplayMonth(newDate)
            }}
          >
            <SelectTrigger className="h-8 w-[110px] text-sm">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {Array.from({ length: 12 }, (_, i) => {
                const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
                // Check if month is within limits
                const year = displayMonth.getFullYear()
                if (fromDate && year === fromDate.getFullYear() && i < fromDate.getMonth()) return null
                if (toDate && year === toDate.getFullYear() && i > toDate.getMonth()) return null
                return (
                  <SelectItem key={i} value={i.toString()}>
                    {monthNames[i]}
                  </SelectItem>
                )
              })}
            </SelectContent>
          </Select>
          
          <Select 
            value={displayMonth.getFullYear().toString()} 
            onValueChange={(value) => {
              const newDate = new Date(displayMonth)
              newDate.setFullYear(parseInt(value))
              // Adjust month if needed
              if (fromDate && newDate < fromDate) {
                newDate.setMonth(fromDate.getMonth())
              } else if (toDate && newDate > toDate) {
                newDate.setMonth(toDate.getMonth())
              }
              setDisplayMonth(newDate)
            }}
          >
            <SelectTrigger className="h-8 w-[80px] text-sm">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {(() => {
                const fromYear = fromDate?.getFullYear() || 1900
                const toYear = toDate?.getFullYear() || 2100
                return Array.from(
                  { length: Math.min(201, toYear - fromYear + 1) }, 
                  (_, i) => fromYear + i
                ).map(year => (
                  <SelectItem key={year} value={year.toString()}>
                    {year}
                  </SelectItem>
                ))
              })()}
            </SelectContent>
          </Select>
        </div>
        
        <Calendar
          mode="single"
          selected={date}
          onSelect={handleSelect}
          disabled={disabledDates}
          fromDate={fromDate}
          toDate={toDate}
          month={displayMonth}
          onMonthChange={setDisplayMonth}
          initialFocus
          captionLayout="label"
          hideHead={false}
          className={cn(
            "rounded-md",
            calendarClassName
          )}
          classNames={{
            month_caption: "hidden", // Hide the default header since we have our own dropdowns
            nav: "hidden", // Hide navigation arrows since we have dropdowns
            day: cn(
              "transition-all duration-200"
            ),
            day_button: cn(
              "transition-all duration-200",
              "hover:bg-muted",
              "data-[selected=true]:bg-primary data-[selected=true]:text-primary-foreground",
              "data-[selected=true]:hover:bg-primary data-[selected=true]:hover:text-primary-foreground"
            ),
            button_previous: "hidden", // Hide previous button
            button_next: "hidden", // Hide next button
            caption: cn(
              "transition-all duration-200",
              isCalendarHovered && "text-primary"
            ),
            today: cn(
              "relative",
              "after:absolute after:bottom-1 after:left-1/2 after:-translate-x-1/2",
              "after:h-0.5 after:w-4 after:bg-primary after:rounded-full",
              "after:transition-all after:duration-200"
            ),
          }}
        />
      </PopoverContent>
    </Popover>
  )
}

interface DateRangePickerProps {
  dateRange?: { from: Date | undefined; to: Date | undefined }
  onSelect?: (range: { from: Date | undefined; to: Date | undefined } | undefined) => void
  placeholder?: string
  disabled?: boolean
  buttonClassName?: string
  calendarClassName?: string
  formatStr?: string
  numberOfMonths?: number
  disabledDates?: any
  fromDate?: Date
  toDate?: Date
}

export function DateRangePicker({
  dateRange,
  onSelect,
  placeholder = "Pick a date range",
  disabled = false,
  buttonClassName,
  calendarClassName,
  formatStr = "LLL dd, y",
  numberOfMonths = 2,
  disabledDates,
  fromDate,
  toDate,
}: DateRangePickerProps) {
  const [open, setOpen] = React.useState(false)
  const [isCalendarHovered, setIsCalendarHovered] = React.useState(false)
  const [tempRange, setTempRange] = React.useState<{ from: Date | undefined; to: Date | undefined } | undefined>(dateRange)
  const [displayMonth, setDisplayMonth] = React.useState<Date>(dateRange?.from || new Date())

  // Initialize temp range when popover opens
  React.useEffect(() => {
    if (open) {
      setTempRange(dateRange) // Start with existing dates
    }
  }, [open, dateRange])

  // Update parent when closing with valid selection
  React.useEffect(() => {
    if (!open && tempRange?.from && tempRange?.to && 
        tempRange.from.getTime() !== tempRange.to.getTime()) {
      // Only update parent when closing with both dates selected
      onSelect?.(tempRange)
    }
  }, [open, tempRange, onSelect])

  const handleSelect = (range: any) => {
    // If we already have a complete range (both from and to with different dates)
    if (tempRange?.from && tempRange?.to && 
        tempRange.from.getTime() !== tempRange.to.getTime()) {
      // Reset to just the newly clicked date as start date
      // The range parameter might have both dates, so we need to extract just the clicked date
      let clickedDate: Date | undefined;
      
      // If the new range has the same start date as our temp range, the user clicked a new date
      if (range?.from && range?.to) {
        // Determine which date is the newly clicked one
        if (range.from.getTime() !== tempRange.from.getTime() && 
            range.from.getTime() !== tempRange.to.getTime()) {
          clickedDate = range.from;
        } else if (range.to.getTime() !== tempRange.from.getTime() && 
                   range.to.getTime() !== tempRange.to.getTime()) {
          clickedDate = range.to;
        } else {
          // If both dates match our existing range, use the from date
          clickedDate = range.from;
        }
      } else if (range?.from) {
        clickedDate = range.from;
      }
      
      const newRange = { from: clickedDate, to: undefined }
      setTempRange(newRange)
      return
    }
    
    setTempRange(range)
    
    // Check if we now have a complete range and close if we do
    if (range?.from && range?.to && range.from.getTime() !== range.to.getTime()) {
      // We have both dates selected, close the popover after a small delay
      setTimeout(() => {
        setOpen(false)
      }, 150)
    }
  }

  const displayText = React.useMemo(() => {
    // Always show the actual dateRange (not tempRange) in the button
    if (!dateRange?.from) return placeholder
    if (!dateRange.to) return format(dateRange.from, formatStr)
    return `${format(dateRange.from, formatStr)} - ${format(dateRange.to, formatStr)}`
  }, [dateRange, formatStr, placeholder])

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          disabled={disabled}
          className={cn(
            "w-[300px] justify-start text-left font-normal",
            "",
            "transition-all duration-200",
            "active:scale-100", // Override the default button scale
            !dateRange?.from && "text-muted-foreground",
            disabled && "opacity-50 cursor-not-allowed",
            open && "ring-2 ring-ring ring-offset-2 ring-offset-background",
            buttonClassName
          )}
        >
          <CalendarIcon 
            className={cn(
              "mr-2 h-4 w-4 shrink-0",
              "transition-transform duration-200",
              open && "rotate-12"
            )} 
          />
          <span className="truncate">{displayText}</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent 
        className={cn(
          "w-auto p-0",
          "animate-in fade-in-0",
          "data-[side=bottom]:slide-in-from-top-2",
          "data-[side=left]:slide-in-from-right-2",
          "data-[side=right]:slide-in-from-left-2",
          "data-[side=top]:slide-in-from-bottom-2"
        )}
        align="start"
        onMouseEnter={() => setIsCalendarHovered(true)}
        onMouseLeave={() => setIsCalendarHovered(false)}
      >
        {/* Year/Month Selection - Always visible */}
        <div className="flex items-center justify-between p-3 pb-0">
          <div className="flex items-center gap-2">
            <Select 
              value={displayMonth.getMonth().toString()} 
              onValueChange={(value) => {
                const newDate = new Date(displayMonth)
                newDate.setMonth(parseInt(value))
                setDisplayMonth(newDate)
              }}
            >
              <SelectTrigger className="h-8 w-[110px] text-sm">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {Array.from({ length: 12 }, (_, i) => {
                  const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
                  // Check if month is within limits
                  const year = displayMonth.getFullYear()
                  if (fromDate && year === fromDate.getFullYear() && i < fromDate.getMonth()) return null
                  if (toDate && year === toDate.getFullYear() && i > toDate.getMonth()) return null
                  return (
                    <SelectItem key={i} value={i.toString()}>
                      {monthNames[i]}
                    </SelectItem>
                  )
                })}
              </SelectContent>
            </Select>
            
            <Select 
              value={displayMonth.getFullYear().toString()} 
              onValueChange={(value) => {
                const newDate = new Date(displayMonth)
                newDate.setFullYear(parseInt(value))
                // Adjust month if needed
                if (fromDate && newDate < fromDate) {
                  newDate.setMonth(fromDate.getMonth())
                } else if (toDate && newDate > toDate) {
                  newDate.setMonth(toDate.getMonth())
                }
                setDisplayMonth(newDate)
              }}
            >
              <SelectTrigger className="h-8 w-[80px] text-sm">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {(() => {
                  const fromYear = fromDate?.getFullYear() || 1900
                  const toYear = toDate?.getFullYear() || 2100
                  return Array.from(
                    { length: Math.min(201, toYear - fromYear + 1) }, 
                    (_, i) => fromYear + i
                  ).map(year => (
                    <SelectItem key={year} value={year.toString()}>
                      {year}
                    </SelectItem>
                  ))
                })()}
              </SelectContent>
            </Select>
          </div>
          
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 hover:bg-muted"
            onClick={() => setOpen(false)}
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
        
        <Calendar
          mode="range"
          selected={tempRange}
          onSelect={handleSelect}
          disabled={disabledDates}
          fromDate={fromDate}
          toDate={toDate}
          numberOfMonths={numberOfMonths}
          month={displayMonth}
          onMonthChange={setDisplayMonth}
          initialFocus
          captionLayout="label"
          className={cn(
            "rounded-md",
            calendarClassName
          )}
          classNames={{
            month_caption: "hidden", // Hide the default header since we have our own dropdowns
            nav: "hidden", // Hide navigation arrows since we have dropdowns
            day: cn(
              "transition-all duration-200"
            ),
            day_button: cn(
              "transition-all duration-200",
              "hover:bg-muted",
              "data-[selected=true]:bg-primary data-[selected=true]:text-primary-foreground",
              "data-[selected=true]:hover:bg-primary data-[selected=true]:hover:text-primary-foreground"
            ),
            button_previous: "hidden", // Hide previous button
            button_next: "hidden", // Hide next button
            caption: cn(
              "transition-all duration-200",
              isCalendarHovered && "text-primary"
            ),
            range_middle: cn(
              "transition-all duration-300",
              "[&>button]:hover:scale-105"
            ),
            today: cn(
              "relative",
              "after:absolute after:bottom-1 after:left-1/2 after:-translate-x-1/2",
              "after:h-0.5 after:w-4 after:bg-primary after:rounded-full",
              "after:transition-all after:duration-200"
            ),
          }}
        />
      </PopoverContent>
    </Popover>
  )
}