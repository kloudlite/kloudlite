import { useEffect, useState } from 'react';
import List from '~/console/components/list';

interface IConfigItemComponent {
  items: { [key: string]: string };
  onClick: (selected: string) => void;
}

const ConfigItemComponent = ({
  items,
  onClick = (_: any) => _,
}: IConfigItemComponent) => {
  const [selected, setSelected] = useState('');
  useEffect(() => {
    onClick(selected);
  }, [selected]);

  return (
    <List.Root>
      {Object.entries(items).map(([key, v]) => {
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
                render: () => <div className="bodyMd text-text-soft">{v}</div>,
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

export default ConfigItemComponent;
