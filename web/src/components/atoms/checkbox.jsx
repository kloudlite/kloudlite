import { Switch } from '@headlessui/react'
import PropTypes from "prop-types";
import { BounceIt } from "../bounce-it";
import classNames from "classnames";
import { useEffect, useState } from 'react';


export const Checkbox = ({label, error, ...props})=>{
  const [checked, setChecked] = useState(props.checked)
  useEffect(()=>{
    if(props.onChange) props.onChange(checked)
  }, [checked])
  return <BounceIt disable={props.disabled} className='cursor-pointer'>
      <Switch className={classNames(
        "flex gap-2 items-center group",
        "focus:ring-2 ring-border-focus ring-offset-1",
        "outline-none transition-all rounded",
        )} {...props} checked={checked} onChange={setChecked}>
        {({checked}) => (
          <>
            <div className={
              classNames(
                "border  w-5 h-5 rounded relative outline-none transition-all",
                {
                  "border-border-default": !checked && !error && !props.disabled,
                },
                {
                  "group-hover:bg-surface-hovered": !checked && !error,
                },
                error?{
                  "bg-surface-danger-subdued border-border-danger": !checked,
                }:{
                  "bg-surface-default border-border-default": !checked,
                },
                error?{
                  "bg-surface-danger-default border-border-danger": checked && !props.disabled,
                }:{
                  "bg-surface-primary-default border-border-primary": checked &&!props.disabled,
                },
                {
                  "border-border-disabled": props.disabled,
                }
              )
            }>
            {checked && !props.indeterminate && (
                  <div className="absolute top-2/4 left-2/4 -translate-y-2/4 -translate-x-2/4">
                  <svg width="17" height="16" viewBox="0 0 17 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M14.5 4.00019L6.5 11.9998L2.5 8.00019" className={
                      classNames({
                        "stroke-text-disabled":props.disabled,
                        "stroke-text-on-primary":!props.disabled,
                      })
                    } stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                </div>
                )}
                {props.indeterminate && (
                  <div className="absolute top-2/4 left-2/4 -translate-y-2/4 -translate-x-2/4">
                  <svg width="17" height="16" viewBox="0 0 17 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M2.5 8H14.5" className={
                      classNames({
                        "stroke-text-disabled":props.disabled,
                        "stroke-text-on-primary":!props.disabled,
                      })
                    } stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                </div>
                  
                )}
            </div>
                
            <span className={classNames({
              "text-text-disabled":props.disabled,
              "text-text-default":!props.disabled,
            }, "select-none bodyMd-medium")}>{label}</span>
          </>
        )}
    </Switch>
  </BounceIt>
}

Checkbox.propTypes = {
  label: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  disabled: PropTypes.bool,
  error: PropTypes.bool,
  indeterminate: PropTypes.bool,
  checked: PropTypes.bool,
}

Checkbox.defaultProps = {
  onChange: ()=>{},
  disabled: false,
  error: false,
  indeterminate: false,
}
