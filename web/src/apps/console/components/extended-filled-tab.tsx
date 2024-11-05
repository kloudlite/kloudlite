import { Key } from 'react';
import Tabs, { ITab } from '@kloudlite/design-system/atoms/tabs';
import { cn } from '@kloudlite/design-system/utils';
import { NonNullableString } from '~/root/lib/types/common';
import { Button } from '@kloudlite/design-system/atoms/button';

export interface IExtendedFilledTab<T = string> {
  value: string;
  onChange?: (item: T) => void;
  items: ITab<T>[];
  size?: 'md' | 'sm' | NonNullableString;
}
const ExtendedFilledTab = <T,>({
  value,
  onChange,
  items = [],
  size = 'md',
}: IExtendedFilledTab<T>) => {
  return (
    <div
      className={cn('bg-surface-basic-active rounded inline-block w-fit', {
        'p-lg shadow-button border border-border-default': size === 'md',
        'p-md': size === 'sm',
      })}
    >
      <Tabs.Root size="sm" variant="filled" value={value} onChange={onChange}>
        {items.map((item) => (
          <Tabs.Tab
            key={item.value as Key}
            label={item.label}
            value={item.value}
            prefix={item.prefix}
          />
        ))}
      </Tabs.Root>
    </div>
  );
};

export default ExtendedFilledTab;
