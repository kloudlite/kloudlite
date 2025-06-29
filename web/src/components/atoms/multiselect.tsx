"use client";

import * as React from "react";
import { X, ChevronDown, Check } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/molecules/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/molecules/popover";
import { Badge } from "@/components/atoms/badge";

export interface MultiSelectOption {
  label: string;
  value: string;
  disabled?: boolean;
}

interface MultiSelectProps {
  options: MultiSelectOption[];
  selected: string[];
  onChange: (values: string[]) => void;
  placeholder?: string;
  searchPlaceholder?: string;
  emptyText?: string;
  disabled?: boolean;
  className?: string;
  maxItems?: number;
  enableSearch?: boolean;
  error?: boolean;
}

export function MultiSelect({
  options,
  selected,
  onChange,
  placeholder = "Select items...",
  searchPlaceholder = "Search...",
  emptyText = "No items found.",
  disabled = false,
  className,
  maxItems = 3,
  enableSearch = false,
  error,
}: MultiSelectProps) {
  const [open, setOpen] = React.useState(false);
  const triggerRef = React.useRef<HTMLButtonElement>(null);

  const handleSelect = (value: string) => {
    const newSelected = selected.includes(value)
      ? selected.filter((item) => item !== value)
      : [...selected, value];
    onChange(newSelected);
  };

  const handleRemove = (value: string, e: React.MouseEvent) => {
    e.stopPropagation();
    onChange(selected.filter((item) => item !== value));
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLButtonElement>) => {
    if (!open && (e.key === "ArrowDown" || e.key === "ArrowUp" || e.key === " ")) {
      e.preventDefault();
      setOpen(true);
    }
  };

  const selectedOptions = options.filter((option) =>
    selected.includes(option.value)
  );

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
          ref={triggerRef}
          disabled={disabled}
          onKeyDown={handleKeyDown}
          aria-invalid={error}
          className={cn(
            baseStyles,
            error ? errorStyles : normalStyles,
            className
          )}
        >
          <div className="flex flex-1 items-center gap-1 overflow-hidden">
            {selectedOptions.length === 0 ? (
              <span className="text-muted-foreground">{placeholder}</span>
            ) : (
              <>
                {selectedOptions.slice(0, maxItems).map((option) => (
                  <Badge
                    key={option.value}
                    variant="secondary"
                    className="h-5 gap-0.5 px-1.5 pr-0.5 flex-shrink-0"
                  >
                    <span className="text-xs truncate max-w-[100px]">{option.label}</span>
                    <div
                      onClick={(e) => handleRemove(option.value, e)}
                      className="ml-0.5 rounded-full outline-none hover:bg-muted focus:bg-muted p-0.5 cursor-pointer"
                      role="button"
                      tabIndex={0}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault();
                          handleRemove(option.value, e);
                        }
                      }}
                    >
                      <X className="h-2.5 w-2.5" />
                    </div>
                  </Badge>
                ))}
                {selectedOptions.length > maxItems && (
                  <Badge variant="secondary" className="h-5 px-1.5 flex-shrink-0">
                    <span className="text-xs">
                      +{selectedOptions.length - maxItems} more
                    </span>
                  </Badge>
                )}
              </>
            )}
          </div>
          <ChevronDown className="h-4 w-4 opacity-50 ml-2 shrink-0" />
        </button>
      </PopoverTrigger>
      <PopoverContent 
        className="w-[var(--radix-popover-trigger-width)] p-0" 
        align="start"
        sideOffset={4}
      >
        <Command>
          {enableSearch && (
            <CommandInput placeholder={searchPlaceholder} autoFocus />
          )}
          <CommandEmpty>{emptyText}</CommandEmpty>
          <CommandList>
            <CommandGroup>
              {options.map((option) => {
                const isSelected = selected.includes(option.value);
                return (
                  <CommandItem
                    key={option.value}
                    value={option.value}
                    onSelect={() => handleSelect(option.value)}
                    disabled={option.disabled}
                    className="relative cursor-pointer py-1.5 pl-2 pr-8"
                  >
                    <span className="absolute right-2 flex h-3.5 w-3.5 items-center justify-center">
                      {isSelected && <Check className="h-4 w-4" />}
                    </span>
                    <span>{option.label}</span>
                  </CommandItem>
                );
              })}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}