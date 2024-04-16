import { DotsThreeVerticalFill } from '@jengaicons/react';
import { Link } from '@remix-run/react';
import { Key, useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import OptionList from '~/components/atoms/option-list';

export type IResourceExtraItem =
  | {
      to?: string;
      linkProps?: {
        target?: string;
        rel?: string;
      };
      icon?: JSX.Element;
      prefix?: JSX.Element;
      suffix?: JSX.Element;
      key: Key;
      label: string;
      type: 'item';
      className?: string;
      onClick?: () => void;
    }
  | { type: 'separator'; key: Key };

interface IResourceExtraAction {
  options: Array<IResourceExtraItem>;
  disabled?: boolean;
}

const ResourceExtraAction = ({
  options = [],
  disabled,
}: IResourceExtraAction) => {
  const [open, setOpen] = useState(false);

  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <IconButton
          disabled={disabled}
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
                {...item.linkProps}
                LinkComponent={Link}
                key={`${item.key}-extra-item-option`}
                className={item.className}
                onClick={item.onClick}
              >
                {(item.icon && item.icon) || (item.prefix && item.prefix)}
                {item.label}
                {item.suffix && item.suffix}
              </OptionList.Link>
            );
          }

          return (
            <OptionList.Item
              key={`${item.key}-extra-item-option`}
              className={item.className}
              onClick={item.onClick}
            >
              {(item.icon && item.icon) || (item.prefix && item.prefix)}
              {item.label}
              {item.suffix && item.suffix}
            </OptionList.Item>
          );
        })}
      </OptionList.Content>
    </OptionList.Root>
  );
};
export default ResourceExtraAction;
