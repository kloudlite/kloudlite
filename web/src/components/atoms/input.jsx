import classNames from "classnames"
import PropTypes from "prop-types";
import { useEffect, useState } from "react";
import { useNumberFieldState } from "react-stately";
import { useLocale } from "react-aria";
import { useNumberField, useButton } from "react-aria";
import { CaretUpFill, CaretDownFill, Search, XCircleFill } from "@jengaicons/react";
import { Pressable } from '@ark-ui/react';
import { BounceIt } from "../bounce-it.jsx";
import { useRef } from "react";

const Button = ({ className, ...props }) => {
  let ref = useRef(null);
  let { buttonProps } = useButton(props, ref);
  return <button {...buttonProps} ref={ref} className={className}>{props.children}</button>;
}

export const NumberInput = ({ infoContent, error, message, ...props }) => {
  const { locale } = useLocale();
  const inputRef = useRef();
  const state = useNumberFieldState({ ...props, locale });
  let {
    labelProps,
    inputProps,
    groupProps,
    incrementButtonProps,
    decrementButtonProps
  } = useNumberField(props, state, inputRef);
  return <div className="flex flex-col gap-1">
    <div className="flex">
      <label {...labelProps} className="flex-1 select-none bodyMd-medium">{props.label}</label>
      {infoContent && <div className="bodyMd">{infoContent}</div>}
    </div>
    <div className="flex relative" {...groupProps}>
      <input
        {...inputProps}
        ref={inputRef}
        className={classNames(
          "transition-all outline-none flex-1",
          "border",
          "rounded px-3 py-2 bodyMd ",
          "ring-offset-1 focus-visible:ring-2 focus:ring-border-focus",
          {
            "bg-surface-danger-subdued border-border-danger text-text-danger placeholder:text-critical-400": error,
            "text-text-default border-border-default": !error
          }
        )}
      />
      <div className="flex flex-col absolute right-0 top-0 bottom-0 justify-center items-center border-l border-border-default divide-border-default divide-y">
        <Button {...incrementButtonProps} className={"flex-1 p-0.5"} ><CaretUpFill size={16} /></Button>
        <Button {...decrementButtonProps} className={"flex-1 p-0.5"}><CaretDownFill size={16} /></Button>
      </div>
    </div>

    {message && <span className={classNames("bodySm", {
      "text-text-danger": error,
      "text-text-default": !error
    })}>{message}</span>}
  </div>
}

export const TextInput = ({ label, disabled, infoContent, placeholder, value = '', onChange, error, message, Component = "input", Prefix, Postfix, showclear, className }) => {
  const [val, setVal] = useState(value)
  useEffect(() => {
    if (onChange) {
      onChange(val)
    }
  }, [val])
  const C = Component || "input"
  return <div className={classNames("flex flex-col",
    {
      "gap-1": label || infoContent
    },
    className
  )}>
    <div className="flex">
      <label className="flex-1 select-none bodyMd-medium">{label}</label>
      {infoContent && <div className="bodyMd">{infoContent}</div>}
    </div>
    <div className={(classNames("transition-all px-3 rounded border flex flex-row items-center relative ring-offset-1 focus-within:ring-2 focus-within:ring-border-focus",
      {
        "text-text-danger bg-surface-danger-subdued border-border-danger": error,
        "text-text-default border-border-default": !error,
        "text-text-disabled border-border-disabled bg-surface-input": disabled
      }))}>
      {Prefix && C === "input" && <div className={classNames("pr-2 bodyMd",
        {
          "text-text-strong": typeof (Prefix) !== "object" && !error && !disabled,
          "text-text-danger": error,
          "text-text-disabled":disabled
        })}>{typeof (Prefix) === "object" ? <Prefix size={20} color="currentColor" /> : Prefix}</div>}
      <C
        value={val}
        onChange={(e) => {
          setVal(e.target.value)
        }}
        disabled={disabled}
        placeholder={placeholder}
        className={classNames(
          "outline-none disabled:bg-surface-input disabled:text-text-disabled",
          "w-full",
          "rounded py-2 bodyMd ",
          "",
          {
            "text-text-danger bg-surface-danger-subdued placeholder:text-critical-400": error,
            "text-text-default": !error
          }
        )}
      />
      {Postfix && <div className={classNames("pl-2 bodyMd",
        {
          "text-text-danger": error,
          "text-text-strong": !error && !disabled,
          "text-text-disabled": disabled
        })}>{typeof (Postfix) === "object" ? <Postfix size={20} color="currentColor" /> : Postfix}</div>}
      {showclear && <BounceIt className="pl-2" disable={disabled}>
        <Pressable disabled={disabled} className='outline-none flex items-center rounded ring-offset-1 focus-visible:ring-2 focus:ring-border-focus justify-center' onPress={(e) => {
          setVal('')
        }}>
          <XCircleFill size={20} color="currentColor" />
        </Pressable>
      </BounceIt>}
    </div>
    {message && <span className={classNames("bodySm", {
      "text-text-danger": error,
      "text-text-default": !error
    })}>{message}</span>}
  </div>
}

TextInput.propTypes = {
  label: PropTypes.string,
  placeholder: PropTypes.string,
  value: PropTypes.string,
  onChange: PropTypes.func,
  error: PropTypes.bool,
  message: PropTypes.string,
  infoContent: PropTypes.elementType,
  className: PropTypes.string,
  Component: PropTypes.string
}

TextInput.defaultProps = {
  placeholder: "Placeholder",
  value: "",
  onChange: () => { },
}