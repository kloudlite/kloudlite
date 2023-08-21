import { DotsThreeVerticalFill, Trash } from '@jengaicons/react';
import { useEffect, useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import { TextArea } from '~/components/atoms/input';
import OptionList from '~/components/atoms/option-list';
import { cn } from '~/components/utils';
import List from '~/console/components/list';

const Resources = ({ items = [], modifiedItems = [], setModifiedData }) => {
  useEffect(() => {
    console.log(
      modifiedItems,
      !modifiedItems.find((mi) => mi.key === 'DATABASE_URL')
    );
  }, [modifiedItems]);
  return (
    <List.Root>
      {items.map((item) => (
        <List.Item
          key={item.key}
          items={[
            {
              key: 1,
              className: 'flex-1',
              render: () => (
                <div className="flex flex-row items-center gap-3xl">
                  <div
                    className={cn(
                      'bodyMd-semibold text-text-default w-[300px]',
                      {
                        '!text-text-critical line-through': !modifiedItems.find(
                          (mi) => mi.key === item.key
                        ),
                      }
                    )}
                  >
                    {item.key}
                  </div>
                  <div
                    className={cn('bodyMd text-text-soft flex-1', {
                      '!text-text-critical line-through': !modifiedItems.find(
                        (mi) => mi.key === item.key
                      ),
                    })}
                  >
                    {item.value}
                  </div>
                  <ResourceItemExtraOptions
                    onDelete={() => {
                      setModifiedData(
                        modifiedItems.filter((mi) => mi.key !== item.key)
                      );
                    }}
                  />
                </div>
              ),
            },
          ]}
        />
      ))}
    </List.Root>
  );
};

const ResourceItemExtraOptions = ({ onDelete = null }) => {
  const [open, setOpen] = useState(false);
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <IconButton
          variant="plain"
          icon={DotsThreeVerticalFill}
          selected={open}
          onClick={(e) => {
            e.stopPropagation();
          }}
          onMouseDown={(e) => {
            e.stopPropagation();
          }}
          onPointerDown={(e) => {
            e.stopPropagation();
          }}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.Item className="!text-text-critical" onSelect={onDelete}>
          <Trash size={16} />
          <span>Delete</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList.Root>
  );
};

export default Resources;
