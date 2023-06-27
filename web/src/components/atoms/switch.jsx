import PropTypes from "prop-types";
import classNames from "classnames";
import { useState, useEffect, useMemo, useRef, useId } from 'react';
import { motion } from "framer-motion"
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
          "data-[state=unchecked]:bg-surface-default data-[state=unchecked]:border-border-default",
          "data-[state=checked]:bg-surface-primary-default data-[state=checked]:border-border-primary",
          {
            "bg-surface-default border-border-disabled !cursor-default": props.disabled
          }
        )}
        id={id}
        disabled={props.disabled}
        onCheckedChange={(e) => { setChecked(e) }}
      >
        <SW.Thumb className={classNames(
          "w-5.25 h-5.25 rounded-full translate-x-0 transition-all duration-200 data-[state=checked]:translate-x-full",
          "",
          {
            'bg-icon-disabled': props.disabled,
            'data-[state=unchecked]:bg-surface-primary-default data-[state=checked]:bg-surface-default': !props.disabled,
          }
        )} />
      </SW.Root>
      {props.label && <label
        className={classNames({
          "text-text-disabled": props.disabled,
          "text-text-default cursor-pointer": !props.disabled,
        }, "bodyMd-medium")}
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
