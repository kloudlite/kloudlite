import PropTypes from "prop-types";
import { RadioGroup as RadioGroupComp } from '@headlessui/react'
import { useEffect, useState } from 'react';
import classNames from "classnames";

export const RadioGroup = ({ label, items, onChange, className, ...props }) => {
  const [value, setValue] = useState(props.value)
  useEffect(() => {
    if (onChange) onChange(value)
  }, [value])
  return <RadioGroupComp value={value} onChange={setValue} className={classNames("flex flex-col", className)}>
    <RadioGroupComp.Label>{label}</RadioGroupComp.Label>
    {
      items.map((item) => {
        return (
          <RadioGroupComp.Option key={item.key} value={item.value} className={classNames("group  outline-none w-fit", {
            "pointer-events-none": item.disabled,
          })} disabled={item.disabled}>
            {({ checked, disabled }) => (
              <div className={
                classNames("flex group gap-2 items-center cursor-pointer")
              }>
                <div className={classNames("w-5 h-5 rounded-full border group-hover:bg-surface-hovered group-focus:ring-2 ring-border-focus ring-offset-1 transition-all flex items-center justify-center",
                  disabled ? {
                    "border-border-disabled": true,
                  } : {
                    "border-border-default": !checked,
                    "border-border-primary": checked,
                  },
                )}>
                  {checked && (<div className={classNames(
                    "block w-3 h-3  rounded-full",
                    {
                      "bg-surface-disabled-default": disabled,
                      "bg-surface-primary-default": !disabled,
                    },
                  )}></div>)}
                </div>
                <div className={classNames({
                  "text-text-disabled": disabled,
                  "text-text-default": !disabled,
                }, "bodyMd-medium")}>{item.label}</div>
              </div>
            )}
          </RadioGroupComp.Option>
        )
      })
    }
  </RadioGroupComp>
}

RadioGroup.propTypes = {
  items: PropTypes.arrayOf(PropTypes.shape({
    label: PropTypes.string,
    value: PropTypes.oneOfType([PropTypes.string, PropTypes.object]),
    key: PropTypes.string
  })).isRequired,
  label: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  disabled: PropTypes.bool,
  error: PropTypes.bool,
  checked: PropTypes.bool,
}

RadioGroup.defaultProps = {
  label: "item",
  onChange: () => { },
  disabled: false,
  error: false,
}
