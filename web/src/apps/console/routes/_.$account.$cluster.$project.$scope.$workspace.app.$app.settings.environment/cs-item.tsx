import { useEffect, useState } from 'react';
import List from '~/console/components/list';

interface ICSComponent {
  items: { [key: string]: string };
  type: 'config' | 'secret';
  onClick: (selected: string) => void;
}

const CSComponent = ({
  items,
  type,
  onClick = (_: any) => _,
}: ICSComponent) => {
  const [selected, setSelected] = useState('');
  useEffect(() => {
    onClick(selected);
  }, [selected]);

  return (
    <List.Root>
      {Object.entries(items).map(([key, v]) => {
        return (
          <List.Row
            key={key}
            pressed={selected === key}
            onClick={() => {
              setSelected((prev) => (prev === key ? '' : key));
            }}
            columns={[
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
                  <div className="bodyMd text-text-soft">
                    {type === 'config' ? v : '•••••••••••••••'}
                  </div>
                ),
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

export default CSComponent;
