import classNames from "classnames"
import PropTypes, { object, string } from "prop-types";
import { cloneElement, useEffect, useState } from "react";
import { useNumberFieldState } from "react-stately";
import { useLocale } from "react-aria";
import { useNumberField, useButton } from "react-aria";
import { CaretUpFill, CaretDownFill, XCircleFill } from "@jengaicons/react";
import { useRef } from "react";
import { useTextField } from "react-aria";


const Button = ({ className, ...props }) => {
  let ref = useRef(null);
  let { buttonProps } = useButton(props, ref);
  return <button {...buttonProps} ref={ref} className={className}>{props.children}</button>;
}

export const NumberInput = (props = { label, disabled, message, extra, placeholder, value: '', onChange, error: false, className }) => {
  const { locale } = useLocale();
  const inputRef = useRef();
  const state = useNumberFieldState({ isDisabled: props.disabled, ...props, locale });

  let {
    labelProps,
    inputProps,
    groupProps,
    incrementButtonProps,
    decrementButtonProps,
  } = useNumberField(props, state, inputRef);

  return <div className={classNames("flex flex-col",
    {
      "gap-1": props.label || props.extra
    })}>
    <div className="flex">
      <label {...labelProps} className="flex-1 select-none bodyMd-medium">{props.label}</label>
      {props.extra && <div className="bodyMd">{cloneElement(props.extra)}</div>}
    </div>
    <div className="flex relative" {...groupProps}>
      <input
        {...inputProps}
        ref={inputRef}
        disabled={props.disabled}
        className={classNames(
          "outline-none flex-1",
          "border",
          "outline-none disabled:bg-surface-input disabled:text-text-disabled",
          "rounded px-3 py-2 bodyMd ",
          "ring-offset-1 focus-visible:ring-2 focus:ring-border-focus",
          {
            "bg-surface-danger-subdued border-border-danger text-text-danger placeholder:text-critical-400": props.error,
            "text-text-default border-border-default": !props.error
          }
        )}
      />
      <div className={classNames("flex flex-col absolute right-px top-px bottom-px justify-center items-center",
        {
          "bg-surface-danger-subdued divide-border-danger divide-y rounded-r border-l border-border-danger text-text-danger placeholder:text-critical-400": props.error,
          "text-text-default border-l border-border-default divide-border-default divide-y": !props.error
        })}>
        <Button {...incrementButtonProps} className={classNames("flex-1 p-0.5 disabled:text-icon-disabled")} ><CaretUpFill size={16} color="currentColor" /></Button>
        <Button {...decrementButtonProps} className={classNames("flex-1 p-0.5 disabled:text-icon-disabled")}><CaretDownFill size={16} color="currentColor" /></Button>
      </div>
    </div>

    {props.message && <span className={classNames("bodySm", {
      "text-text-danger": props.error,
      "text-text-default": !props.error
    })}>{props.message}</span>}
  </div>
}

export const TextInputBase = (props = { label, disabled, message, extra, placeholder, value: '', onChange, error: false, prefix, suffix, showclear, className, component }) => {

  const [val, setVal] = useState(props.value ? props.value : '')
  useEffect(() => {
    if (props.onChange) {
      props.onChange(val)
    }
  }, [val])

  let ref = useRef(null);
  let { labelProps, inputProps, errorMessageProps } = useTextField({
    ...props, isDisabled: props.disabled, errorMessage: props.message,
    onChange: (e) => {
      setVal(e)
    },
    value: val
  }, ref);


  const Prefix = props.prefix
  const Suffix = props.suffix

  const Component = props.component || "input"

  return (
    <div className={classNames("flex flex-col",
      {
        "gap-1": props.label || props.extra
      },
      props.className
    )}>
      <div className="flex">
        <label {...labelProps} className="flex-1 select-none bodyMd-medium">{props.label}</label>
        {props.extra && <div className="bodyMd">{cloneElement(props.extra)}</div>}
      </div>
      <div className={(classNames("px-3 rounded border flex flex-row items-center relative ring-offset-1 focus-within:ring-2 focus-within:ring-border-focus",
        {
          "text-text-danger bg-surface-danger-subdued border-border-danger": props.error,
          "text-text-default border-border-default": !props.error,
          "text-text-disabled border-border-disabled bg-surface-input": props.disabled,
          "pr-0": props.component != "input"
        }))}>
        {Prefix && <div className={classNames("pr-2 bodyMd",
          {
            "text-text-strong": typeof (Prefix) !== "object" && !props.error && !props.disabled,
            "text-text-danger": props.error,
            "text-text-disabled": props.disabled
          })}>{typeof (Prefix) === "object" ? <Prefix size={20} color="currentColor" /> : Prefix}</div>}
        <Component {...inputProps} ref={ref} className={classNames(
          "outline-none disabled:bg-surface-input disabled:text-text-disabled",
          "w-full",
          "rounded py-2 bodyMd ",
          {
            "text-text-danger bg-surface-danger-subdued placeholder:text-critical-400": props.error,
            "text-text-default": !props.error
          }
        )} />
        {Suffix && <div className={classNames("pl-2 bodyMd",
          {
            "text-text-danger": props.error,
            "text-text-strong": !props.error && !props.disabled,
            "text-text-disabled": props.disabled
          })}>{typeof (Suffix) === "object" ? <Suffix size={20} color="currentColor" /> : Suffix}</div>}
        {
          props.showclear && <Button
            onPress={() => {
              setVal('')
            }}
            className={classNames('outline-none flex items-center rounded ring-offset-1 focus-visible:ring-2 focus:ring-border-focus justify-center',
              {
                "cursor-default": props.disabled
              })}>
            <XCircleFill size={20} color="currentColor" />
          </Button>}
      </div>

      {props.message && (
        <div {...errorMessageProps} className={classNames("bodySm", {
          "text-text-danger": props.error,
          "text-text-default": !props.error
        })}>
          {props.message}
        </div>
      )}
    </div>
  );
}

export const TextInput = (props = { label, disabled, extra, placeholder, value: '', onChange, error, message, prefix, suffix, showclear, className }) => {
  return <TextInputBase {...props} component={'input'} />
}

export const TextArea = (props = { label, disabled, extra, placeholder, value: '', onChange, error, message, className }) => {
  return <TextInputBase {...props} component={'textarea'} />
}



const BaseInputProps = {
  label: PropTypes.string,
  placeholder: PropTypes.string,
  value: PropTypes.string,
  onChange: PropTypes.func,
  error: PropTypes.bool,
  message: PropTypes.string,
  extra: PropTypes.elementType,
  className: PropTypes.string,
  disabled: PropTypes.bool,
}

TextInput.propTypes = {
  ...BaseInputProps,
  prefix: PropTypes.oneOfType([string, object]),
  suffix: PropTypes.oneOfType([string, object])
}

TextInput.propTypes = {
  ...BaseInputProps
}

NumberInput.propTypes = {
  ...BaseInputProps
}

TextInput.defaultProps = {
  placeholder: "Placeholder",
  value: "",
  disabled: false,
  onChange: () => { },
}