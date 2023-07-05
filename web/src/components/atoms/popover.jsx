import * as PopoverPrimitive from '@radix-ui/react-popover'
import { Command } from 'cmdk'
import { useState } from 'react'

export const Popover = ({ }) => {
    const [value, setValue] = useState('apple')
    return (
        <PopoverPrimitive.Root>
            <PopoverPrimitive.Trigger>Toggle PopoverPrimitive</PopoverPrimitive.Trigger>

            <PopoverPrimitive.Content>
                <Command shouldFilter={false} value={value} onValueChange={setValue}>
                    <Command.Input />
                    <Command.List>
                        <Command.Item>Apple</Command.Item>
                        <Command.Item>Orange</Command.Item>
                    </Command.List>
                </Command>
            </PopoverPrimitive.Content>
        </PopoverPrimitive.Root>
    )
}