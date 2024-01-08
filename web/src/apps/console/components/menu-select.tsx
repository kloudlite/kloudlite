import React, { ReactNode, useState } from 'react';
import * as Select from '@radix-ui/react-select';
import { cn } from '~/components/utils';
import { AnimatePresence, motion } from 'framer-motion';

interface ISelectItem {
  children: ReactNode;
  className?: string;
  value: string;
  active?: boolean;
}

const SelectItem = React.forwardRef<HTMLDivElement, ISelectItem>(
  ({ children, className, ...props }, forwardedRef) => {
    return (
      <Select.Item
        onClick={(e) => {
          console.log(e);
        }}
        className={cn(
          'group relative flex flex-row gap-xl items-center bodyMd gap cursor-pointer select-none py-lg px-xl text-text-default outline-none transition-colors focus:bg-surface-basic-hovered hover:bg-surface-basic-hovered data-[disabled]:pointer-events-none data-[disabled]:text-text-disabled data-[state=checked]:bg-surface-basic-active',
          className
        )}
        {...props}
        ref={forwardedRef}
      >
        <Select.ItemText>{children}</Select.ItemText>
      </Select.Item>
    );
  }
);

interface IMenuSelect {
  trigger: ReactNode;
  value?: string;
  items: {
    label: ReactNode;
    value: string;
  }[];
  onChange?: (value: string) => void;
  onClick?: (value: string) => void;
}
const MenuSelect = ({
  trigger,
  onClick,
  value,
  items,
  onChange,
}: IMenuSelect) => {
  const [open, setOpen] = useState(false);
  return (
    <Select.Root
      open={open}
      onOpenChange={setOpen}
      onValueChange={onChange}
      value={value}
    >
      <Select.Trigger>
        <Select.Value>{trigger}</Select.Value>
      </Select.Trigger>
      <AnimatePresence>
        {open && (
          <Select.Portal>
            <Select.Content asChild>
              <motion.div
                initial={{ opacity: 0, scale: 0.85 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.85 }}
                transition={{ duration: 0.3, ease: 'anticipate' }}
                className={cn(
                  'z-50 border border-border-default shadow-popover bg-surface-basic-default rounded min-w-[160px] overflow-hidden origin-top py-lg'
                )}
              >
                <Select.Viewport
                  onClick={(e) => {
                    console.log(e);
                  }}
                >
                  {items.map((item) => (
                    <div key={item.value} onClick={() => onClick?.(item.value)}>
                      <SelectItem key={item.value} value={item.value}>
                        <div
                          onClick={() => {
                            console.log('clicked');
                          }}
                        >
                          {item.label}
                        </div>
                      </SelectItem>
                    </div>
                  ))}
                </Select.Viewport>
              </motion.div>
            </Select.Content>
          </Select.Portal>
        )}
      </AnimatePresence>
    </Select.Root>
  );
};

export default MenuSelect;
