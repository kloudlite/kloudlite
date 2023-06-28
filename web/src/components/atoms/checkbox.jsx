import PropTypes from "prop-types";
import classNames from "classnames";
import { useEffect, useId, useState } from 'react';

import * as CB from '@radix-ui/react-checkbox';

export const Checkbox = (props) => {
  const [checked, setChecked] = useState(props.checked)
  useEffect(() => {
    if (props.onChange) props.onChange(checked)
  }, [checked])
  const id = useId();
  return (
    <div className="flex items-center justify-center w-fit">
      <CB.Root
        className={classNames("rounded flex flex-row items-center justify-center border w-5 h-5 outline-none transition-all cursor-pointer",
          "ring-border-focus ring-offset-1 focus:ring-2",
          {
            "border-border-disabled !cursor-default": props.disabled,
          },
          {
            "bg-surface-default border-border-default": !checked && !props.disabled,
            "bg-surface-primary-default border-border-primary": checked && !props.error && !props.disabled,
            "bg-surface-danger-default border-border-danger": checked && props.error && !props.disabled,
            "hover:bg-surface-hovered": !checked && !props.disabled
          })}
        defaultChecked
        id={id}
        checked={checked}
        onCheckedChange={(e) => { setChecked((prev) => props.indeterminate ? prev == "indeterminate" ? false : "indeterminate" : e) }}
        disabled={props.disabled}

      >
        <CB.Indicator>
          <svg width="17" height="16" viewBox="0 0 17 16" fill="none" xmlns="http://www.w3.org/2000/svg">
            {
              checked == true && <path d="M14.5 4.00019L6.5 11.9998L2.5 8.00019" className={
                classNames({
                  "stroke-text-disabled": props.disabled,
                  "stroke-text-on-primary": !props.disabled,

                })
              } strokeLinecap="round" strokeLinejoin="round" />
            }
            {
              checked === "indeterminate" && <path d="M2.5 8H14.5" className={
                classNames({
                  "stroke-text-disabled": props.disabled,
                  "stroke-text-on-primary": !props.disabled,
                })
              } strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
            }
          </svg>
        </CB.Indicator>
      </CB.Root>
      {
        props.label && <label
          className={classNames({
            "text-text-disabled": props.disabled,
            "text-text-default cursor-pointer": !props.disabled,
          }, "bodyMd-medium pl-2 select-none")}
          htmlFor={id}>
          {props.label}
        </label>
      }
    </div >
  );
}

Checkbox.propTypes = {
  label: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  disabled: PropTypes.bool,
  error: PropTypes.bool,
  checked: PropTypes.any,
  indeterminate: PropTypes.bool
}

Checkbox.defaultProps = {
  onChange: () => { },
  disabled: false,
  error: false,
}
