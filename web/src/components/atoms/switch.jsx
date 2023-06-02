import { Switch as SwitchComp } from '@headlessui/react';
import PropTypes from "prop-types";
import { BounceIt } from "../bounce-it";
import classNames from "classnames";
import { useState, useEffect, useMemo } from 'react';
import { motion } from "framer-motion"

export const Switch = ({ error, disabled, ...props }) => {
  const [checked, setChecked] = useState(props.checked);
  const layoutId = useMemo(() => Math.random().toString(36).substring(2), [])

  useEffect(() => {
    if (props.onChange) props.onChange(checked)
  }, [checked])

  return <BounceIt disable={props.disabled} className='cursor-pointer'>
    <SwitchComp className={classNames(
      "flex gap-1 items-center group",
      "focus:ring-2 ring-border-focus ring-offset-1",
      "outline-none transition-all rounded-full",
    )} {...props} checked={checked} onChange={setChecked} disabled={disabled}>
      {({ checked }) => (
        <>
          <div className={classNames(
            'w-[48px] rounded-full border  flex items-center p-0.5  transition-all',
            disabled ? 'bg-surface-default border-border-disabled' : {
              'bg-surface-primary-default border-border-primary': checked,
              'bg-surface-default border-border-default': !checked,
            }
          )}>
            <div className='w-[21px] h-[21px]'>
              {
                checked ? null : (<motion.div layoutId={layoutId} className={
                  classNames(
                    'rounded-full w-full h-full',
                    {
                      'bg-icon-disabled': disabled,
                      'bg-surface-primary-default': !disabled,
                    }
                  )
                } />)
              }

            </div>
            <div className='w-[21px] h-[21px]'>
              {
                checked ? (<motion.div layoutId={layoutId} className={
                  classNames(
                    ' rounded-full  w-full h-full',
                    {
                      'bg-icon-disabled': disabled,
                      'bg-surface-default': !disabled,
                    }
                  )
                } />) : null
              }
            </div>
          </div>
        </>
      )}
    </SwitchComp>
  </BounceIt>
}

Switch.propTypes = {
  label: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  disabled: PropTypes.bool,
  error: PropTypes.bool,
  checked: PropTypes.bool,
}

Switch.defaultProps = {
  onChange: () => { },
  disabled: false,
  error: false,
}
