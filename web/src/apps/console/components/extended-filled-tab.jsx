import Tabs from '~/components/atoms/tabs';

const ExtendedFilledTab = ({ value, onChange, items = [] }) => {
  return (
    <div className="bg-surface-basic-active rounded border border-border-default shadow-button inline-block p-lg w-fit">
      <Tabs.Root size="sm" variant="filled" value={value} onChange={onChange}>
        {items.map((item) => (
          <Tabs.Tab
            key={item.to}
            label={item.label}
            value={item.to}
            to={item.to}
          />
        ))}
      </Tabs.Root>
    </div>
  );
};

export default ExtendedFilledTab;
