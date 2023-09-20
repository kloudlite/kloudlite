import { DotsThreeVerticalFill } from '@jengaicons/react';
import { Key, useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import OptionList from '~/components/atoms/option-list';

interface IResourceExtraItem {
  icon: JSX.Element;
  key: Key;
  label: string;
  type: 'separator' | 'item';
  className?: string;
  onClick?: () => void;
}
interface IResourceExtraAction {
  options: Array<IResourceExtraItem>;
}

const ResourceExtraAction = ({ options = [] }: IResourceExtraAction) => {
  const [open, setOpen] = useState(false);

  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <IconButton
          variant="plain"
          icon={<DotsThreeVerticalFill />}
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
        {options.map((item) => {
          if (item.type === 'separator') {
            return <OptionList.Separator />;
          }
          return (
            <OptionList.Item
              key={`${item.label}-extra-item-option`}
              className={item.className}
              onClick={item.onClick}
            >
              {item.icon && item.icon}
              {item.label}
            </OptionList.Item>
          );
        })}
      </OptionList.Content>
    </OptionList.Root>
  );
};
export default ResourceExtraAction;
