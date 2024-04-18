import { ReactNode, useEffect, useState } from 'react';
import AnimateHide from '~/components/atoms/animate-hide';
import { Button, IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { cn, uuid } from '~/components/utils';
import { MinusCircle, Plus } from '~/iotconsole/components/icons';

interface IKeyValuePair {
  onChange?(
    itemArray: Array<Record<string, any>>,
    itemObject: Record<string, any>
  ): void;
  value?: Array<Record<string, any>>;
  label?: ReactNode;
  message?: ReactNode;
  error?: boolean;
  selectMessage?: ReactNode;
  selectError?: boolean;
  selectLoading?: boolean;
  size?: 'lg' | 'md';
  addText?: string;
  options: { label: string; value: string; updateInfo: null }[];
  keyLabel?: string;
  valueLabel?: string;
  regexPath?: RegExp;
}
const KeyValuePairSelect = ({
  onChange,
  value = [],
  label,
  message,
  error,
  selectMessage,
  selectError,
  selectLoading,
  size,
  addText,
  options,
  keyLabel = 'key',
  valueLabel = 'value',
  regexPath,
}: IKeyValuePair) => {
  const newItem = [{ [keyLabel]: '', [valueLabel]: '', id: uuid() }];
  const [items, setItems] = useState<Array<Record<string, any>>>(newItem);

  const handleChange = (_value = '', id = '', target = {}) => {
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
    if (onChange) onChange(Array.from(tempItems), formatItems);
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

  const [isValidPath, setIsValidPath] = useState<{
    [key: string]: boolean;
  }>();

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
                <Select
                  creatable
                  size={size || 'md'}
                  value={item[keyLabel]}
                  options={async () => options}
                  onChange={(_, val) => {
                    console.log('val', val);
                    handleChange(val, item.id, 'key');
                  }}
                  error={selectError}
                  message={selectMessage}
                  loading={selectLoading}
                  disableWhileLoading
                />
              </div>
              <div className="flex-1">
                <TextInput
                  size={size || 'md'}
                  error={isValidPath?.[item.id] === false}
                  message={
                    isValidPath?.[item.id] === false ? 'Invalid path' : ''
                  }
                  placeholder="Value"
                  value={item[valueLabel]}
                  onChange={({ target }) => {
                    if (target.value === '') {
                      setIsValidPath({ ...isValidPath, [item.id]: true });
                    } else if (regexPath?.test(target.value)) {
                      setIsValidPath({ ...isValidPath, [item.id]: true });
                    } else {
                      setIsValidPath({ ...isValidPath, [item.id]: false });
                    }
                    handleChange(target.value, item.id, 'value');
                  }}
                />
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

export default KeyValuePairSelect;
