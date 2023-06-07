import PropTypes from "prop-types";
import { BounceIt }from "../bounce-it";
import classNames from "classnames";
import { useEffect, useRef, useState } from 'react';


import { useFocusRing, VisuallyHidden } from 'react-aria';
import { useToggleState } from "react-stately";
import { useCheckbox } from "react-aria";

export const Checkbox = (props) => {
  const [checked, setChecked] = useState(props.checked)
  useEffect(() => {
    if (props.onChange) props.onChange(checked)
  }, [checked])

  let state = useToggleState({
    ...props, isDisabled: props.disabled, isSelected: checked, onChange: (e) => {
      setChecked(e)
    }
  });
  let ref = useRef(null);
  let { inputProps } = useCheckbox({ ...props, isIndeterminate: props.indeterminate, isDisabled: props.disabled }, state, ref);
  let { isFocusVisible, focusProps } = useFocusRing();
  let isSelected = state.isSelected && !props.indeterminate;

  return (
    <label
      className="flex items-center gap-x-2 justify-center w-fit"
    >
      <VisuallyHidden className="peer">
        <input {...inputProps} {...focusProps} ref={ref} />
      </VisuallyHidden>

      <div className={classNames("rounded flex flex-row items-center justify-center border w-5 h-5 outline-none transition-all cursor-pointer ring-border-focus ring-offset-1",
        {
          "border-border-default": !isSelected && !props.error && !inputProps.disabled,
        },
        {
          "hover:bg-surface-hovered": (!state.isSelected) && !props.error,
        },
        {
          "ring-2": isFocusVisible,
        },
        props.error ? {
          "bg-surface-danger-subdued border-border-danger": !state.isSelected,
        } : {
          "bg-surface-default border-border-default": !state.isSelected,
        },
        props.error ? {
          "bg-surface-danger-default border-border-danger": state.isSelected && !inputProps.disabled,
        } : {
          "bg-surface-primary-default border-border-primary": (state.isSelected) && !inputProps.disabled,
        },
        {
          "border-border-disabled !cursor-default": inputProps.disabled,
        },
      )}>
        <svg width="17" height="16" viewBox="0 0 17 16" fill="none" xmlns="http://www.w3.org/2000/svg">
          {
            isSelected && <path d="M14.5 4.00019L6.5 11.9998L2.5 8.00019" className={
              classNames({
                "stroke-text-disabled": inputProps.disabled,
                "stroke-text-on-primary": !inputProps.disabled,

              })
            } strokeLinecap="round" strokeLinejoin="round" />
          }
          {
            state.isSelected && props.indeterminate && <path d="M2.5 8H14.5" className={
              classNames({
                "stroke-text-disabled": inputProps.disabled,
                "stroke-text-on-primary": !inputProps.disabled,
              })
            } strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
          }
        </svg>

      </div>
      {props.label && <span className={classNames({
        "text-text-disabled cursor-default": inputProps.disabled,
        "text-text-default cursor-pointer": !inputProps.disabled,
      }, "select-none bodyMd-medium")}>{props.label}</span>}
    </label>
  );
}

Checkbox.propTypes = {
  label: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  disabled: PropTypes.bool,
  error: PropTypes.bool,
  indeterminate: PropTypes.bool,
  checked: PropTypes.bool
}

Checkbox.defaultProps = {
  onChange: () => { },
  disabled: false,
  error: false,
  indeterminate: false,
}
