import PropTypes from "prop-types";
import { useEffect, useId, useState } from 'react';
import classNames from "classnames";
import * as RG from '@radix-ui/react-radio-group';


const RadioGroupItem = (props) => {
  let id = useId()
  return <div className={classNames("flex items-center w-fit",
    {
      "cursor-pointer": !props.disabled
    })}>
    <RG.Item
      className={classNames("w-5 h-5 outline-none rounded-full border ring-border-focus ring-offset-1 focus:ring-2 transition-all flex items-center justify-center border-border-default",
        {
          "hover:bg-surface-hovered": !props.disabled,
          "data-[state=checked]:border-border-primary data-[disabled]:border-border-disabled": true
        })}
      value={props.value}
      id={id}
      disabled={props.disabled}
    >
      <RG.Indicator
        className={classNames(
          "block w-3 h-3  rounded-full",
          {
            "bg-icon-disabled": props.disabled,
            "bg-surface-primary-default": !props.disabled,
          },
        )}
      />
    </RG.Item>
    {props.label && <label
      className={classNames({
        "text-text-disabled": props.disabled,
        "text-text-default cursor-pointer": !props.disabled,
      }, "bodyMd-medium pl-2")}
      htmlFor={id}>
      {props.label}
    </label>}
  </div>
}


export const RadioGroup = (props) => {
  const [value, setValue] = useState(props.value)
  useEffect(() => {
    if (props.onChange) props.onChange(value)
  }, [value])
  return <RG.Root className="flex flex-col gap-y-2.5" value={value} aria-label={props.label} disabled={props.disabled}
    onValueChange={(e) => {
      setValue(e);
      console.log(e);
    }}>
    <span className="bodyMd-medium">{props.label}</span>
    {props.items && props.items.map((item) => <RadioGroupItem label={item.label} value={item.value} disabled={item.disabled || props.disabled} key={item.key} />)}
  </RG.Root>
}

RadioGroup.propTypes = {
  items: PropTypes.arrayOf(PropTypes.shape({
    label: PropTypes.string,
    value: PropTypes.oneOfType([PropTypes.string, PropTypes.object]),
    disabled: PropTypes.bool,
    key: PropTypes.string
  })).isRequired,
  label: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  error: PropTypes.bool,
  checked: PropTypes.bool,
  disabled: PropTypes.bool,
}

RadioGroup.defaultProps = {
  label: "item",
  onChange: () => { },
  error: false,
  disabled: false
}
