import { DotsThreeVerticalFill } from '@jengaicons/react';
import { Link } from '@remix-run/react';
import { Key, useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import OptionList from '~/components/atoms/option-list';

type IResourceExtraItem =
  | {
      to?: string;
      icon: JSX.Element;
      key: Key;
      label: string;
      type: 'item';
      className?: string;
      onClick?: () => void;
    }
  | { type: 'separator'; key: Key };

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
          size="sm"
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
            return <OptionList.Separator key={item.key} />;
          }
          if (item.to) {
            return (
              <OptionList.Link
                to={item.to}
                LinkComponent={Link}
                key={`${item.key}-extra-item-option`}
                className={item.className}
                onClick={item.onClick}
              >
                {item.icon && item.icon}
                {item.label}
              </OptionList.Link>
            );
          }

          return (
            <OptionList.Item
              key={`${item.key}-extra-item-option`}
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
