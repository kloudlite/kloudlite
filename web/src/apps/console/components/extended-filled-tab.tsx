import Tabs, { ITab } from '~/components/atoms/tabs';
import { cn } from '~/components/utils';
import { NonNullableString } from '~/root/lib/types/common';

interface IExtendedFilledTab {
  value: string;
  onChange?: (item: string) => void;
  items: ITab[];
  size?: 'md' | 'sm' | NonNullableString;
}
const ExtendedFilledTab = ({
  value,
  onChange,
  items = [],
  size = 'md',
}: IExtendedFilledTab) => {
  return (
    <div
      className={cn(
        'bg-surface-basic-active rounded border border-border-default inline-block w-fit',
        {
          'p-lg shadow-button': size === 'md',
          'p-md': size === 'sm',
        }
      )}
    >
      <Tabs.Root size="sm" variant="filled" value={value} onChange={onChange}>
        {items.map((item) => (
          <Tabs.Tab
            key={item.value}
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
