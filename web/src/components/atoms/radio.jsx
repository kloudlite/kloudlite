import PropTypes from "prop-types";
import { createContext, useContext, useEffect, useRef, useState } from 'react';
import classNames from "classnames";
import { useFocusRing, VisuallyHidden, useRadio } from "react-aria";
import { useRadioGroupState } from "react-stately";
import { useRadioGroup } from "react-aria";


let RadioContext = createContext(null);

export const RadioGroup = (props) => {
  let { items, label, disabled } = props;

  const [value, setValue] = useState(props.value)
  useEffect(() => {
    if (props.onChange) props.onChange(value)
  }, [value])

  let state = useRadioGroupState({
    ...props, value: value, onChange: (e) => {
      setValue(e)
    }, isDisabled: disabled
  });
  let { radioGroupProps, labelProps } =
    useRadioGroup(props, state);


  return (
    <div {...radioGroupProps} className="flex flex-col gap-y-[10px]">
      <span {...labelProps}>{label}</span>
      <RadioContext.Provider value={state}>
        {items && items.map((item) => {
          return <Radio label={item.label} disabled={item.disabled} value={item.value} key={item.key} />
        })}
      </RadioContext.Provider>
    </div>
  );
}

export const Radio = (props) => {
  let state = useContext(RadioContext);
  let ref = useRef(null);
  let { inputProps, isSelected, isDisabled } = useRadio({ ...props, isDisabled: props.disabled, "aria-label": props.label }, state, ref);
  let { isFocusVisible, focusProps } = useFocusRing();

  return (
    <label
      className="flex gap-2 items-center cursor-pointer w-fit group"
    >
      <VisuallyHidden>
        <input {...inputProps} {...focusProps} ref={ref} />
      </VisuallyHidden>
      <div className={classNames("w-5 h-5 rounded-full border group-hover:bg-surface-hovered ring-border-focus ring-offset-1 transition-all flex items-center justify-center",
        isDisabled ? {
          "border-border-disabled": true,
        } : {
          "border-border-default": !isSelected,
          "border-border-primary": isSelected,
        },
        {
          "ring-2": isFocusVisible
        }
      )}>
        {isSelected && (<div className={classNames(
          "block w-3 h-3  rounded-full",
          {
            "bg-surface-disabled-default": isDisabled,
            "bg-surface-primary-default": !isDisabled,
          },
        )}></div>)}
      </div>
      <div className={classNames({
        "text-text-disabled": isDisabled,
        "text-text-default": !isDisabled,
      }, "bodyMd-medium")}>{props.label}</div>
    </label>
  );
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
