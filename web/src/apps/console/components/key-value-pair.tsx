import { ReactNode, useEffect, useState } from 'react';
import AnimateHide from '~/components/atoms/animate-hide';
import { Button, IconButton } from '~/components/atoms/button';
import { NumberInput, TextInput } from '~/components/atoms/input';
import { cn, uuid } from '~/components/utils';
import { MinusCircle, Plus } from '~/console/components/icons';

interface IKeyValuePair {
  onChange?(
    itemArray: Array<Record<string, any>>,
    itemObject: Record<string, any>,
    itemArrayWithoutId: Array<Record<string, any>>
  ): void;
  value?: Array<Record<string, any>>;
  label?: ReactNode;
  message?: ReactNode;
  error?: boolean;
  size?: 'lg' | 'md';
  addText?: string;
  keyLabel?: string;
  valueLabel?: string;
  type?: 'number' | 'text';
}
const KeyValuePair = ({
  onChange,
  value = [],
  label,
  message,
  error,
  size,
  addText,
  keyLabel = 'key',
  valueLabel = 'value',
  type = 'text',
}: IKeyValuePair) => {
  const newItem = [{ [keyLabel]: '', [valueLabel]: '', id: uuid() }];
  const [items, setItems] = useState<Array<Record<string, any>>>(newItem);

  const handleChange = (
    _value: string | number,
    id: string | number,
    target = {}
  ) => {
    console.log(typeof _value, id);

    const tempItems = items.map((i) => {
      if (i.id === id) {
        switch (target) {
          case 'key':
            return { ...i, [keyLabel]: _value };

          case 'value':
          default:
            return { ...i, [valueLabel]: _value };
        }
      }
      return i;
    });

    const formatItems = tempItems.reduce((acc, curr) => {
      if (curr.key && curr.value) {
        acc[curr.key] = curr.value;
      }
      return acc;
    }, {});

    const x = JSON.parse(JSON.stringify(tempItems));
    const withoutId = x.map((i: any) => {
      delete i.id;
      return i;
    });

    if (onChange) onChange(Array.from(tempItems), formatItems, withoutId);
  };

  useEffect(() => {
    if (value && value.length === 0) {
      setItems(newItem);
      return;
    }
    setItems(
      Array.from(value || newItem).map((v) => ({
        ...v,
        id: v.id ? v.id : uuid(),
      }))
    );
  }, [value]);

  return (
    <div className="flex flex-col">
      <div className="flex flex-col">
        <div className="flex flex-col gap-md">
          {label && (
            <span className="text-text-default bodyMd-medium">{label}</span>
          )}
          {items.map((item) => (
            <div key={item.id} className="flex flex-row gap-xl items-start">
              <div className="flex-1">
                {type === 'text' && (
                  <TextInput
                    size={size || 'md'}
                    error={error}
                    placeholder="Key"
                    value={item[keyLabel]}
                    onChange={({ target }) =>
                      handleChange(target.value, item.id, 'key')
                    }
                  />
                )}
                {type === 'number' && (
                  <NumberInput
                    size={size || 'md'}
                    error={error}
                    placeholder="Key"
                    value={item[keyLabel]}
                    onChange={({ target }) =>
                      handleChange(parseInt(target.value, 10), item.id, 'key')
                    }
                  />
                )}
              </div>
              <div className="flex-1">
                {type === 'text' && (
                  <TextInput
                    size={size || 'md'}
                    error={error}
                    placeholder="Value"
                    value={item[valueLabel]}
                    onChange={({ target }) =>
                      handleChange(target.value, item.id, 'value')
                    }
                  />
                )}
                {type === 'number' && (
                  <NumberInput
                    size={size || 'md'}
                    error={error}
                    placeholder="Value"
                    value={item[valueLabel]}
                    onChange={({ target }) =>
                      handleChange(parseInt(target.value, 10), item.id, 'value')
                    }
                  />
                )}
              </div>
              <div className="self-center">
                <IconButton
                  icon={<MinusCircle />}
                  variant="plain"
                  disabled={items.length < 2}
                  onClick={() => {
                    setItems(items.filter((i) => i.id !== item.id));
                  }}
                />
              </div>
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
            content={addText || 'Add'}
            size="sm"
            prefix={<Plus />}
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
