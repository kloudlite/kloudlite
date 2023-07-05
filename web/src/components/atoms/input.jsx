import classNames from "classnames"
import PropTypes from "prop-types";
import React, { cloneElement, forwardRef, useEffect, useId, useState } from "react";
import { CaretUpFill, CaretDownFill, XCircleFill, EyeSlash, Eye } from "@jengaicons/react";
import { useRef } from "react";


export const NumberInput = (props = { label, disabled, message, extra, placeholder, value: '', onChange, error: false, className, step, min, max }) => {

  const [v, setV] = useState(props.value || props.min || 0)
  const ref = useRef();
  const id = useId();

  const step = props.step || 1;

  useEffect(() => {
    if (props.onChange) {
      props.onChange({
        target: {
          ...ref.current,
          value: v,
        },
      })
    }

  }, [v])

  return <div className={classNames("flex flex-col",
    {
      "gap-1": props.label || props.extra
    })}>
    <div className="flex">
      <label className="flex-1 select-none bodyMd-medium" htmlFor={id}>{props.label}</label>
      {props.extra && <div className="bodyMd">{cloneElement(props.extra)}</div>}
    </div>
    <div className={classNames("transition-all flex relative", "ring-offset-1 focus-within:ring-2 focus-within:ring-border-focus rounded border overflow-hidden",
      {
        "bg-surface-danger-subdued border-border-danger text-text-danger placeholder:text-critical-400": props.error,
        "text-text-default border-border-default": !props.error
      }
    )}>
      <input
        ref={ref}
        id={id}
        disabled={props.disabled}
        type="number"
        autoComplete="off"
        inputMode="numeric"
        className={classNames(
          "outline-none flex-1",
          "outline-none disabled:bg-surface-input disabled:text-text-disabled",
          "rounded px-3 py-2 bodyMd ",
          "no-spinner"
        )}
        value={v}
        onChange={(e) => {
          setV(e.target.value);
        }}
        min={props.min}
        max={props.max}
      />
      <div className={classNames("flex flex-col absolute right-0 top-0 bottom-0 justify-center items-center",
        {
          "bg-surface-danger-subdued divide-border-danger divide-y rounded-r border-l border-border-danger text-text-danger placeholder:text-critical-400": props.error,
          "text-text-default border-l border-border-default divide-border-default divide-y": !props.error
        })}
        tabIndex={-1}>
        <button
          aria-controls={id}
          aria-label={`Increase ${props.label}`}
          tabIndex={-1}

          onClick={() => {
            setV((v) => v + step);
            ref.current.focus();
          }} className={classNames("flex-1 p-0.5 disabled:text-icon-disabled hover:bg-surface-hovered active:bg-surface-pressed")} ><CaretUpFill size={16} color="currentColor" /></button>
        <button
          aria-controls={id}
          aria-label={`Decrease ${props.label}`}
          tabIndex={-1}
          onClick={() => {
            setV((v) => v - step);
          }} className={classNames("flex-1 p-0.5 disabled:text-icon-disabled hover:bg-surface-hovered active:bg-surface-pressed")}><CaretDownFill size={16} color="currentColor" /></button>
      </div>
    </div>

    {props.message && <span className={classNames("bodySm", {
      "text-text-danger": props.error,
      "text-text-default": !props.error
    })}>{props.message}</span>}
  </div>
}

export const TextInputBase = forwardRef((props = { label, disabled, message, extra, placeholder, value: '', onChange, error: false, prefix, suffix, showclear, className, component, type, onKeyDown, autoComplete }, ref) => {

  const [val, setVal] = useState(props.value || '')
  const [type, setType] = useState(props.type || "text")


  let id = useId()

  useEffect(() => {
    if (props.onChange) {
      props.onChange({
        target: {
          ...ref.current,
          value: val,
        },
      })
    }
  }, [val])



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
      <div className="flex items-center">
        <label className="flex-1 select-none bodyMd-medium" htmlFor={id}>{props.label}</label>
        <div className={classNames(
          {
            "h-6": props.label || props.extra
          }
        )}>{props.extra && cloneElement(props.extra)}</div>
      </div>
      <div className={(classNames("transition-all px-3 rounded border flex flex-row items-center relative ring-offset-1 focus-within:ring-2 focus-within:ring-border-focus ",
        {
          "text-text-danger bg-surface-danger-subdued border-border-danger": props.error,
          "text-text-default border-border-default bg-surface-input": !props.error,
          "text-text-disabled border-border-disabled bg-surface-input": props.disabled,
          "pr-0": props.component != "input"
        }))}>
        {Prefix && <div className={classNames("pr-2 bodyMd",
          {
            "text-text-strong": typeof (Prefix) !== "object" && !props.error && !props.disabled,
            "text-text-danger": props.error,
            "text-text-disabled": props.disabled
          })}>{typeof (Prefix) === "object" ? <Prefix size={20} color="currentColor" /> : Prefix}</div>}
        <Component
          type={type}
          placeholder={props.placeholder}
          id={id}
          className={classNames(
            "outline-none disabled:bg-surface-input disabled:text-text-disabled",
            "w-full",
            "rounded py-2 bodyMd ",
            {
              "text-text-danger bg-surface-danger-subdued placeholder:text-critical-400": props.error,
              "text-text-default bg-surface-input": !props.error
            }
          )}
          value={val}
          onChange={(e) => { setVal(e.target.value) }}
          disabled={props.disabled}
          ref={ref}
          onKeyDown={props.onKeyDown}
          autoComplete={props.autoComplete}
        />
        {Suffix && <div className={classNames("pl-2 bodyMd",
          {
            "text-text-danger": props.error,
            "text-text-strong": !props.error && !props.disabled,
            "text-text-disabled": props.disabled
          })}>{typeof (Suffix) === "object" ? <Suffix size={20} color="currentColor" /> : Suffix}</div>}
        {
          props.showclear && !props.disabled && <button
            tabIndex={-1}
            onClick={() => {
              setVal('')
              console.log('button');
            }}
            className={classNames('outline-none flex items-center rounded justify-center',
              {
                "cursor-default": props.disabled
              })}>
            <XCircleFill size={20} color="currentColor" />
          </button>
        }
        {
          props.type === "password" && !props.disabled && <button
            tabIndex={-1}
            onClick={() => {
              setType((prev) => prev == "text" ? "password" : "text")
              console.log('button');
            }}
            className={classNames('outline-none flex items-center rounded justify-center',
              {
                "cursor-default": props.disabled
              })}>
            {type === "password" ? <EyeSlash size={16} color="currentColor" /> : <Eye size={16} color="currentColor" />}
          </button>
        }
      </div>

      {props.message && (
        <div className={classNames("bodySm", {
          "text-text-danger": props.error,
          "text-text-default": !props.error
        })}>
          {props.message}
        </div>
      )}
    </div>
  );
})

export const TextInput = (props = { label, disabled, extra, placeholder, value: '', onChange, error, message, prefix, suffix, showclear, className }) => {
  let ref = useRef(null)
  return <TextInputBase {...props} component={'input'} ref={ref} type="text" />
}

export const TextArea = (props = { label, disabled, extra, placeholder, value: '', onChange, error, message, className }) => {
  let ref = useRef(null)
  return <TextInputBase {...props} component={'textarea'} ref={ref} type="text" />
}

export const PasswordInput = (props = { label, disabled, extra, placeholder, value: '', onChange, error, message, className }) => {
  let ref = useRef(null)
  return <TextInputBase {...props} component={'input'} ref={ref} type="password" />
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
  prefix: PropTypes.oneOfType([PropTypes.string, PropTypes.object]),
  suffix: PropTypes.oneOfType([PropTypes.string, PropTypes.object])
}

TextInput.propTypes = {
  ...BaseInputProps
}

NumberInput.propTypes = {
  ...BaseInputProps
}

TextInput.defaultProps = {
  placeholder: "",
  value: "",
  disabled: false,
}