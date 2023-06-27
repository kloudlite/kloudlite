import * as React from "react"
import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu"
import classNames from "classnames"
import { CircleFill } from "@jengaicons/react"
import { TextInput, TextInputBase } from "./input"

const DropdownMenu = DropdownMenuPrimitive.Root

const DropdownMenuTrigger = DropdownMenuPrimitive.Trigger

const DropdownMenuGroup = DropdownMenuPrimitive.Group

const DropdownMenuPortal = DropdownMenuPrimitive.Portal

const DropdownMenuSub = DropdownMenuPrimitive.Sub

const DropdownMenuRadioGroup = DropdownMenuPrimitive.RadioGroup

const DropdownMenuContent = React.forwardRef(({ className, sideOffset = 4, ...props }, ref) => (
    <DropdownMenuPrimitive.Portal>
        <DropdownMenuPrimitive.Content
            ref={ref}
            sideOffset={sideOffset}
            align="end"
            className={classNames(
                "z-50 border border-border-default shadow-popover bg-surface-default rounded min-w-48 overflow-hidden",
                "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
                className
            )}
            {...props}
        />
    </DropdownMenuPrimitive.Portal>
))
DropdownMenuContent.displayName = DropdownMenuPrimitive.Content.displayName

const DropdownMenuItem = React.forwardRef(({ className, inset, ...props }, ref) => (
    <DropdownMenuPrimitive.Item
        ref={ref}
        className={classNames(
            "relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
            className
        )}
        {...props}
    />
))
DropdownMenuItem.displayName = DropdownMenuPrimitive.Item.displayName

const DropdownMenuSearchItem = React.forwardRef(({ className, inset, ...props }, ref) => {
    let searchRef = React.useRef(null)
    const setSearchFocus = (e) => {
        e?.preventDefault()
        searchRef.current.focus()
        // console.log(searchRef.current.hasFocus());
    }


    return (
        <DropdownMenuPrimitive.Item
            tabIndex={0}
            ref={ref}
            className={classNames(
                "group relative flex flex-row gap-2.5 items-center bodyMd gap cursor-default select-none py-2 px-3 text-text-default outline-none transition-colors focus:bg-surface-hovered data-[disabled]:pointer-events-none data-[disabled]:text-text-disabled",
                className
            )}
            {...props}
            onSelect={setSearchFocus}
            onClick={(e) => setSearchFocus()}
            onPointerUp={setSearchFocus}
            onPointerDown={(e) => setSearchFocus()}
            onMouseMove={(e) => e.preventDefault()}
            onMouseEnter={(e) => e.preventDefault()}
            onMouseLeave={(e) => e.preventDefault()}
            // onPointerEnter={(e) => e.preventDefault()}
            onPointerMove={(e) => e.preventDefault()}
            onPointerLeave={(e) => e.preventDefault()}


        >
            <TextInputBase component={'input'} ref={searchRef} />
        </DropdownMenuPrimitive.Item>
    )
})
DropdownMenuSearchItem.displayName = DropdownMenuPrimitive.Item.displayName

const DropdownMenuCheckboxItem = React.forwardRef(({ className, children, checked, ...props }, ref) => (
    <DropdownMenuPrimitive.CheckboxItem
        ref={ref}
        className={classNames(
            "group relative flex flex-row gap-2.5 items-center bodyMd gap cursor-default select-none py-2 px-3 text-text-default outline-none transition-colors focus:bg-surface-hovered data-[disabled]:pointer-events-none data-[disabled]:text-text-disabled",
            className
        )}
        checked={checked}
        {...props}
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
DropdownMenuCheckboxItem.displayName =
    DropdownMenuPrimitive.CheckboxItem.displayName

const DropdownMenuRadioItem = React.forwardRef(({ className, children, ...props }, ref) => (
    <DropdownMenuPrimitive.RadioItem
        tabIndex={0}
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
DropdownMenuRadioItem.displayName = DropdownMenuPrimitive.RadioItem.displayName

const DropdownMenuLabel = React.forwardRef(({ className, inset, ...props }, ref) => (
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
DropdownMenuLabel.displayName = DropdownMenuPrimitive.Label.displayName

const DropdownMenuSeparator = React.forwardRef(({ className, ...props }, ref) => (
    <DropdownMenuPrimitive.Separator
        ref={ref}
        className={classNames("-mx-1 my-1 h-px bg-muted", className)}
        {...props}
    />
))
DropdownMenuSeparator.displayName = DropdownMenuPrimitive.Separator.displayName

const DropdownMenuShortcut = ({
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
DropdownMenuShortcut.displayName = "DropdownMenuShortcut"


export const Dropdown = ({ }) => {
    const [position, setPosition] = React.useState("bottom")
    const [showStatusBar, setShowStatusBar] = React.useState(true)
    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <div>
                    <button variant="outline">Open</button>
                    <button variant="outline">Open</button>
                </div>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56">
                <DropdownMenuSearchItem>
                    Hello
                </DropdownMenuSearchItem>
                <DropdownMenuRadioGroup value={position} onValueChange={setPosition} >
                    <DropdownMenuRadioItem value="top" onSelect={(e) => e.preventDefault()}>Top</DropdownMenuRadioItem>
                    <DropdownMenuRadioItem value="bottom" onSelect={(e) => e.preventDefault()}>Bottom</DropdownMenuRadioItem>
                    <DropdownMenuRadioItem value="right" onSelect={(e) => e.preventDefault()}>Right</DropdownMenuRadioItem>
                </DropdownMenuRadioGroup>
                <DropdownMenuCheckboxItem
                    checked={showStatusBar}
                    onCheckedChange={setShowStatusBar}
                    onSelect={(e) => e.preventDefault()}

                >
                    Status Bar
                </DropdownMenuCheckboxItem>
            </DropdownMenuContent>
        </DropdownMenu>
    )
}