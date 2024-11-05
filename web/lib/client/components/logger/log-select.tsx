import { ReactNode } from 'react';
import SelectZener from '@oshq/react-select';
import type {
  IGroupRender,
  IMenuItemRender,
  ISelect,
} from '@oshq/react-select';
import { cn } from '@kloudlite/design-system/utils';
import AnimateHide from '@kloudlite/design-system/atoms/animate-hide';

const menuItemRender = (props: IMenuItemRender) => {
  const { innerProps, render, active, focused } = props;
  return (
    <div
      {...innerProps}
      className={cn('cursor-pointer select-none', {
        'bg-surface-basic-hovered': !!focused && !active,
        'bg-surface-success-pressed': !!active,
      })}
    >
      {typeof render === 'string'
        ? render
        : render?.({ active: !!active, focused: !!focused })}
    </div>
  );
};

const groupRender = ({ label }: IGroupRender) => {
  return (
    <div className="bodySm-medium text-text-disabled px-lg py-md">{label}</div>
  );
};

const Select = <T, U extends boolean | undefined = undefined>(
  props: ISelect<T, U> & {
    label?: ReactNode;
    message?: ReactNode;
    error?: boolean;
  }
) => {
  const {
    value,
    options,
    label,
    placeholder,
    message,
    error = false,
    onChange,
    disabled,
    valueRender,
    creatable,
    multiple,
    onSearch,
    searchable,
    showclear,
    noOptionMessage,
    open,
    disableWhileLoading,
  } = props;

  return (
    <div className="flex flex-col min-w-10xl">
      <div className="flex flex-col gap-md">
        {label && (
          <div className="bodyMd-medium text-text-default h-4xl">{label}</div>
        )}
        <div className="pulsable">
          <div className="pulsable pulsable-hidden">
            <SelectZener
              className={() => {
                const c = cn(
                  'rounded flex flex-row items-center border outline-none cursor-default p-0 bodySm',
                  error && !disabled
                    ? 'bg-surface-critical-subdued border-text-critical text-text-critical'
                    : ''
                );
                return {
                  default: `${c} border-none hljs`,
                  disabled: `${c} border-border-disabled text-text-disabled`,
                  focus: `${c} border-border-default text-text-default ring-offset-1 ring-2 ring-border-focus`,
                };
              }}
              open={open}
              // container bg
              menuClass="shadow-popover hljs"
              menuItemRender={menuItemRender}
              value={value}
              options={options}
              placeholder={
                <div
                  className={cn(
                    error && !disabled
                      ? 'text-text-critical/70'
                      : 'text-text-disabled'
                  )}
                >
                  {placeholder}
                </div>
              }
              showclear={showclear}
              // suffixRender={({ clear, showclear }) =>
              //   suffixRender({
              //     loading: loading || false,
              //     clear,
              //     showclear,
              //     error,
              //     disabled: !!disabled,
              //   })
              // }
              onChange={onChange}
              groupRender={groupRender}
              disabled={disabled}
              valueRender={valueRender}
              creatable={creatable}
              multiple={multiple}
              onSearch={onSearch}
              searchable={searchable}
              noOptionMessage={noOptionMessage}
              disableWhileLoading={disableWhileLoading}
            />
          </div>
        </div>
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
    </div>
  );
};

export default Select;
