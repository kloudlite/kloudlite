import PropTypes from "prop-types";
import classNames from "classnames";
import { useState, useEffect, useId } from 'react';
import * as SW from '@radix-ui/react-switch';

export const Switch = (props) => {

  const [checked, setChecked] = useState(props.checked);
  const id = useId()

  useEffect(() => {
    if (props.onChange) props.onChange(checked)
  }, [checked])

  return (
    <div className='flex gap-2 items-center w-fit'>
      <SW.Root
        className={classNames(
          "transition-all w-12 rounded-full border flex items-center p-0.5 cursor-pointer",
          "ring-border-focus ring-offset-1 focus:ring-2",
          {
            "data-[state=unchecked]:bg-surface-default data-[state=unchecked]:border-border-default": !props.disabled
          },
          {
            "data-[state=checked]:bg-surface-primary-default data-[state=checked]:border-border-primary": !props.disabled
          },
          {
            "data-[disabled]:bg-surface-default data-[disabled]:border-border-disabled data-[disabled]:!cursor-default": props.disabled
          }
        )}
        id={id}
        disabled={props.disabled}
        onCheckedChange={(e) => { setChecked(e) }}
      >
        <SW.Thumb
          className={classNames(
            "group rounded-full translate-x-0 transition-all duration-200 data-[state=checked]:translate-x-full",
          )}
        >
          <svg width="21" height="21" viewBox="0 0 22 22" fill="none" xmlns="http://www.w3.org/2000/svg" className={classNames(
            {
              "group-data-[disabled]:fill-icon-disabled": props.disabled
            },
            {
              "group-data-[state=unchecked]:fill-surface-primary-default group-data-[state=checked]:fill-surface-default": !props.disabled
            }
          )}>
            <circle cx="11" cy="11" r="10.5" />
          </svg>

        </SW.Thumb>

      </SW.Root>
      {props.label && <label
        className={classNames({
          "text-text-disabled": props.disabled,
          "text-text-default cursor-pointer": !props.disabled,
        }, "bodyMd-medium select-none")}
        htmlFor={id}>
        {props.label}
      </label>}
    </div>
  );
}




Switch.propTypes = {
  onChange: PropTypes.func,
  disabled: PropTypes.bool,
  checked: PropTypes.bool,
}

Switch.defaultProps = {
  onChange: () => { },
  disabled: false,
}
