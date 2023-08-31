import Tabs, { TabProps } from '~/components/atoms/tabs';

interface IExtendedFilledTab {
  value: string;
  onChange?: (item: string) => void;
  items: TabProps[];
}
const ExtendedFilledTab = ({
  value,
  onChange,
  items = [],
}: IExtendedFilledTab) => {
  return (
    <div className="bg-surface-basic-active rounded border border-border-default shadow-button inline-block p-lg w-fit">
      <Tabs.Root size="sm" variant="filled" value={value} onChange={onChange}>
        {items.map((item) => (
          <Tabs.Tab key={item.value} label={item.label} value={item.value} />
        ))}
      </Tabs.Root>
    </div>
  );
};

export default ExtendedFilledTab;
