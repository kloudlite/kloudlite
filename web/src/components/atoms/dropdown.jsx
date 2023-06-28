import * as React from "react"
import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu"
import classNames from "classnames"
import { CircleFill } from "@jengaicons/react"
import { TextInput, TextInputBase } from "./input"
import { ButtonBase } from "./button"

const OptionMenu = DropdownMenuPrimitive.Root

const OptionMenuTriggerBase = DropdownMenuPrimitive.Trigger


const OptionMenuRadioGroup = DropdownMenuPrimitive.RadioGroup

const OptionMenuTrigger = React.forwardRef(({ ...props }, ref) => (
    <OptionMenuTriggerBase ref={ref} {...props} asChild />

))

const OptionMenuContent = React.forwardRef(({ className, sideOffset = 4, ...props }, ref) => (
    <DropdownMenuPrimitive.Portal>
        <DropdownMenuPrimitive.Content
            ref={ref}
            sideOffset={sideOffset}
            align="end"
            loop
            className={classNames(
                "OptionContent z-50 border border-border-default shadow-popover bg-surface-default rounded min-w-48 overflow-hidden",
                className
            )}
            {...props}
        />
    </DropdownMenuPrimitive.Portal>
))
OptionMenuContent.displayName = DropdownMenuPrimitive.Content.displayName

const OptionMenuItem = React.forwardRef(({ className, inset, ...props }, ref) => (
    <DropdownMenuPrimitive.Item
        ref={ref}
        className={classNames(
            "relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
            className
        )}
        {...props}
    />
))
// OptionMenuItem.displayName = DropdownMenuPrimitive.Item.displayName


const OptionMenuSearchItem = React.forwardRef(({ onChange }, ref) => {
    let searchRef = React.useRef(null)
    const setSearchFocus = (e) => {
        e?.preventDefault()
        searchRef.current.focus()
    }

    return (
        <DropdownMenuPrimitive.Item
            ref={ref}
            className={classNames(
                "filter-dropdown-item group relative flex flex-row gap-2.5 items-center bodyMd gap cursor-default select-none py-2 px-3 text-text-default outline-none transition-colors focus:bg-surface-hovered data-[disabled]:pointer-events-none data-[disabled]:text-text-disabled",
            )}
            onSelect={setSearchFocus}
            onClick={(e) => setSearchFocus()}
            onPointerUp={setSearchFocus}
            onPointerDown={(e) => setSearchFocus()}
            onMouseMove={(e) => e.preventDefault()}
            onMouseEnter={(e) => e.preventDefault()}
            onMouseLeave={(e) => e.preventDefault()}
            onPointerMove={(e) => e.preventDefault()}
            onPointerLeave={(e) => e.preventDefault()}
            onFocus={(e) => {
                searchRef.current.focus()
            }}
        >
            <TextInputBase component={'input'} ref={searchRef} autoComplete={"off"}
                onKeyDown={(e) => {
                    if (e.key === "ArrowDown" || e.key === "Escape") {
                        e.target.blur()
                        let items = Array.from(document.querySelectorAll('[data-radix-collection-item]'))
                        let searchItemIndex = items.findIndex((x) => x.isSameNode(e.target.closest(".filter-dropdown-item")))
                        items.at(searchItemIndex + 1).focus()
                    }

                }} onChange={(e) => { onChange && onChange(e.target.value) }} />
        </DropdownMenuPrimitive.Item>
    )
})
OptionMenuSearchItem.displayName = DropdownMenuPrimitive.Item.displayName

const OptionMenuCheckboxItem = React.forwardRef(({ className, children, checked, onValueChange, ...props }, ref) => (
    <DropdownMenuPrimitive.CheckboxItem
        ref={ref}
        className={classNames(
            "group relative flex flex-row gap-2.5 items-center bodyMd gap cursor-default select-none py-2 px-3 text-text-default outline-none transition-colors focus:bg-surface-hovered data-[disabled]:pointer-events-none data-[disabled]:text-text-disabled",
            className
        )}
        checked={checked}
        {...props}
        onMouseMove={(e) => e.preventDefault()}
        onMouseEnter={(e) => e.preventDefault()}
        onMouseLeave={(e) => e.preventDefault()}
        onPointerLeave={(e) => e.preventDefault()}
        onPointerEnter={(e) => e.preventDefault()}
        onPointerMove={(e) => e.preventDefault()}
        onCheckedChange={onValueChange}
    >
        <span className="w-5 h-5 rounded border transition-all flex items-center justify-center border-border-default group-data-[state=checked]:border-border-primary group-data-[state=checked]:bg-surface-primary-default group-data-[disabled]:border-border-disabled group-data-[disabled]:bg-surface-default ">
            <DropdownMenuPrimitive.ItemIndicator>
                <svg width="17" height="16" viewBox="0 0 17 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                    {
                        <path d="M14.5 4.00019L6.5 11.9998L2.5 8.00019" className={
                            classNames("stroke-text-on-primary group-data-[disabled]:stroke-text-disabled")
                        } strokeLinecap="round" strokeLinejoin="round" />
                    }
                </svg>
            </DropdownMenuPrimitive.ItemIndicator>
        </span>
        {children}
    </DropdownMenuPrimitive.CheckboxItem>
))
OptionMenuCheckboxItem.displayName =
    DropdownMenuPrimitive.CheckboxItem.displayName

const OptionMenuRadioItem = React.forwardRef(({ className, children, ...props }, ref) => (
    <DropdownMenuPrimitive.RadioItem
        ref={ref}
        className={classNames(
            "group relative flex flex-row gap-2.5 items-center bodyMd gap cursor-default select-none py-2 px-3 text-text-default outline-none transition-colors focus:bg-surface-hovered hover:bg-surface-hovered data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
            className
        )}
        {...props}
        onMouseMove={(e) => e.preventDefault()}
        onMouseEnter={(e) => e.preventDefault()}
        onMouseLeave={(e) => e.preventDefault()}
        onPointerLeave={(e) => e.preventDefault()}
        onPointerEnter={(e) => e.preventDefault()}
        onPointerMove={(e) => e.preventDefault()}
    >
        <span className={classNames(
            "w-5 h-5 rounded-full border transition-all flex items-center justify-center border-border-default group-data-[state=checked]:border-border-primary group-data-[disabled]:border-border-disabled",
        )}>
            <DropdownMenuPrimitive.ItemIndicator>
                <div className={classNames(
                    "block w-3 h-3 rounded-full bg-surface-primary-default group-data-[disabled]:bg-icon-disabled",
                )}></div>
            </DropdownMenuPrimitive.ItemIndicator>
        </span>
        {children}
    </DropdownMenuPrimitive.RadioItem>
))
OptionMenuRadioItem.displayName = DropdownMenuPrimitive.RadioItem.displayName

const OptionMenuLabel = React.forwardRef(({ className, inset, ...props }, ref) => (
    <DropdownMenuPrimitive.Label
        ref={ref}
        className={classNames(
            "px-2 py-1.5 text-sm font-semibold",
            inset && "pl-8",
            className
        )}
        {...props}
    />
))
OptionMenuLabel.displayName = DropdownMenuPrimitive.Label.displayName

const OptionMenuSeparator = React.forwardRef(({ className, ...props }, ref) => (
    <DropdownMenuPrimitive.Separator
        ref={ref}
        className={classNames("h-px bg-border-disabled", className)}
        {...props}
    />
))
OptionMenuSeparator.displayName = DropdownMenuPrimitive.Separator.displayName

const OptionMenuShortcut = ({
    className,
    ...props
}) => {
    return (
        <span
            className={classNames("ml-auto text-xs tracking-widest opacity-60", className)}
            {...props}
        />
    )
}
OptionMenuShortcut.displayName = "OptionMenuShortcut"

const OptionListBase = ({ ...props }) => {
    const [open, setOpen] = React.useState(false)
    console.log(open);
    return (
        <OptionMenu open={open} onOpenChange={setOpen}>
            <OptionMenuTrigger>
                <ButtonBase {...props.trigger.props} sharpLeft={props.sharpLeft} sharpRight={props.sharpRight} className={props.className} selected={open} />
            </OptionMenuTrigger>
            <OptionMenuContent>
                {props.filter && <OptionMenuSearchItem onChange={props.onFilterChange} />}
                {props.children}
            </OptionMenuContent>
        </OptionMenu>
    )
}

export const OptionList = ({ children, trigger, filter, onFilterChange }) => {
    return (
        <OptionListBase children={children} trigger={trigger} filter={filter} onFilterChange={onFilterChange} />
    )
}

export const OptionListGroup = ({ children }) => {
    return (
        <div className="flex flex-row">
            {
                Array.isArray(children) ? children.map((child, index) => {
                    const sharpRight = index < children.length - 1;
                    const sharpLeft = index > 0;
                    return <OptionListBase {...child.props} sharpLeft={sharpLeft} sharpRight={sharpRight} className={classNames({ "-ml-px": (sharpLeft || sharpRight) })} key={`option-list-group-${index}`} />
                }) : <OptionListBase {...children.props} />
            }
        </div>
    )
}

OptionList.RadioGroup = OptionMenuRadioGroup
OptionList.RadioGroupItem = OptionMenuRadioItem
OptionList.CheckboxItem = OptionMenuCheckboxItem
OptionList.Separator = OptionMenuSeparator
