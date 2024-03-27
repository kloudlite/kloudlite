import { ReactNode, useEffect, useState } from 'react';
import AnimateHide from '~/components/atoms/animate-hide';
import { Button, IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { cn, uuid } from '~/components/utils';
import { MinusCircle, Plus } from '~/console/components/icons';

interface IKeyValuePair {
  onChange?(
    itemArray: Array<Record<string, any>>,
    itemObject: Record<string, any>
  ): void;
  value?: Array<Record<string, any>>;
  label?: ReactNode;
  message?: ReactNode;
  error?: boolean;
  size?: 'lg' | 'md';
  addText?: string;
  options: { label: string; value: string; updateInfo: null }[];
}
const KeyValuePairSelect = ({
  onChange,
  value = [],
  label,
  message,
  error,
  size,
  addText,
  options,
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
                  // label={label}
                  value={item.key}
                  options={async () => options}
                  onChange={(_, val) => {
                    console.log('val', val);
                    handleChange(val, item.id, 'key');
                    // handleChange('cacheKey')(dummyEvent(val));
                  }}
                  error={error}
                  message={message}
                  // loading={digestLoading}
                  disableWhileLoading
                />
              </div>
              <div className="flex-1">
                <TextInput
                  size={size || 'md'}
                  error={!!isValidPath?.[item.id]}
                  message={isValidPath?.[item.id] ? 'Invalid path' : ''}
                  placeholder="Value"
                  value={item.value}
                  onChange={({ target }) => {
                    const pathRegex = /^\/(?:[\w.-]+\/)*(?:[\w.-]*)$/;
                    console.log('target.value', target.value);
                    if (pathRegex.test(target.value)) {
                      setIsValidPath({ ...isValidPath, [item.id]: true });
                      // handleChange('cachePath')(dummyEvent(val));
                    } else {
                      setIsValidPath({ ...isValidPath, [item.id]: false });
                    }
                    handleChange(target.value, item.id, 'value');
                    // handleChange(target.value, item.id, 'value');
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
