import { useEffect, useState } from 'react';
import List from '~/console/components/list';

const ConfigItem = ({ items, onClick = (_) => _ }) => {
  const [selected, setSelected] = useState('');
  useEffect(() => {
    onClick(selected);
  }, [selected]);
  return (
    <List.Root>
      {Object.entries(items).map(([key, value]) => {
        return (
          <List.Item
            key={key}
            pressed={selected === key}
            onClick={() => {
              setSelected((prev) => (prev === key ? '' : key));
            }}
            items={[
              {
                key: 1,
                className: 'w-[300px]',
                render: () => (
                  <div className="bodyMd-semibold text-text-default">{key}</div>
                ),
              },
              {
                key: 2,
                className: 'flex-1',
                render: () => (
                  <div className="bodyMd text-text-soft">{value}</div>
                ),
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

export default ConfigItem;
