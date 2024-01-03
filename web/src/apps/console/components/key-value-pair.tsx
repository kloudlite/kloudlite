import { MinusCircle } from '@jengaicons/react';
import { ReactNode, useEffect, useState } from 'react';
import AnimateHide from '~/components/atoms/animate-hide';
import { Button, IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { cn, uuid } from '~/components/utils';

interface IKeyValuePair {
  onChange?(
    itemArray: Array<Record<string, any>>,
    itemObject: Record<string, any>
  ): void;
  value?: Array<Record<string, any>>;
  label?: ReactNode;
  message?: ReactNode;
  error?: boolean;
}
const KeyValuePair = ({
  onChange,
  value = [],
  label,
  message,
  error,
}: IKeyValuePair) => {
  const newItem = [{ key: '', value: '', id: uuid() }];
  const [items, setItems] = useState<Array<Record<string, any>>>(newItem);

  const handleChange = (_value = '', id = '', target = {}) => {
    setItems(
      items.map((i) => {
        if (i.id === id) {
          switch (target) {
            case 'key':
              return { ...i, key: _value };
            case 'value':
            default:
              return { ...i, value: _value };
          }
        }
        return i;
      })
    );
  };

  useEffect(() => {
    const formatItems = items.reduce((acc, curr) => {
      if (curr.key && curr.value) {
        acc[curr.key] = curr.value;
      }
      return acc;
    }, {});
    if (onChange) onChange(Array.from(items), formatItems);
  }, [items]);

  useEffect(() => {
    if (value.length > 0) {
      setItems(Array.from(value).map((v) => ({ ...v, id: uuid() })));
    }
  }, []);

  return (
    <div className="flex flex-col">
      <div className="flex flex-col">
        <div className="flex flex-col gap-md">
          {label && (
            <span className="text-text-default bodyMd-medium">{label}</span>
          )}
          {items.map((item) => (
            <div key={item.id} className="flex flex-row gap-xl items-end">
              <div className="flex-1">
                <TextInput
                  error={error}
                  placeholder="Key"
                  value={item.key}
                  onChange={({ target }) =>
                    handleChange(target.value, item.id, 'key')
                  }
                />
              </div>
              <div className="flex-1">
                <TextInput
                  error={error}
                  placeholder="Value"
                  value={item.value}
                  onChange={({ target }) =>
                    handleChange(target.value, item.id, 'value')
                  }
                />
              </div>
              <IconButton
                icon={<MinusCircle />}
                variant="plain"
                disabled={items.length < 2}
                onClick={() => {
                  setItems(items.filter((i) => i.id !== item.id));
                }}
              />
            </div>
          ))}
        </div>
        <AnimateHide show={!!message}>
          <div
            className={cn(
              'bodySm pulsable',
              {
                'text-text-critical': !!error,
                'text-text-default': !error,
              },
              'pt-md'
            )}
          >
            {message}
          </div>
        </AnimateHide>
        <div className="pt-xl">
          <Button
            variant="basic"
            content="Add arg"
            size="sm"
            onClick={() => {
              setItems([...items, { ...newItem[0], id: uuid() }]);
            }}
          />
        </div>
      </div>
    </div>
  );
};

export default KeyValuePair;
