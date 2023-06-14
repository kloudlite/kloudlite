import PropTypes from "prop-types";
import classNames from "classnames";
import { useState, useEffect, useMemo, useRef } from 'react';
import { motion } from "framer-motion"
import { useToggleState } from 'react-stately';
import { useFocusRing, useSwitch, VisuallyHidden } from 'react-aria';

export const Switch = (props) => {

  const [checked, setChecked] = useState(props.checked);
  const layoutId = useMemo(() => Math.random().toString(36).substring(2), [])

  useEffect(() => {
    if (props.onChange) props.onChange(checked)
  }, [checked])


  let state = useToggleState({
    ...props, isSelected: checked, isDisabled: props.disabled, onChange: (e) => {
      setChecked(e)
    }
  });

  let ref = useRef(null);
  let { inputProps } = useSwitch({ ...props, isDisabled: props.disabled }, state, ref);
  let { isFocusVisible, focusProps } = useFocusRing();

  return (
    <label
      className='flex items-center w-fit'
    >
      <VisuallyHidden>
        <input {...inputProps} {...focusProps} ref={ref} />
      </VisuallyHidden>
      <div className={classNames(
        'transition-all w-12 rounded-full border  flex items-center p-0.5 ring-border-focus ring-offset-1 cursor-pointer',
        props.disabled ? 'bg-surface-default border-border-disabled !cursor-default' : {
          'bg-surface-primary-default border-border-primary': state.isSelected,
          'bg-surface-default border-border-default': !state.isSelected,
        },
        {
          "ring-2": isFocusVisible,
        },

      )}>
        <div className='w-5.25 h-5.25'>
          {
            state.isSelected ? null : (<motion.div layoutId={layoutId} className={
              classNames(
                'rounded-full w-full h-full',
                {
                  'bg-icon-disabled': props.disabled,
                  'bg-surface-primary-default': !props.disabled,
                }
              )
            } />)
          }

        </div>
        <div className='w-5.25 h-5.25'>
          {
            state.isSelected ? (<motion.div layoutId={layoutId} className={
              classNames(
                ' rounded-full  w-full h-full',
                {
                  'bg-icon-disabled': props.disabled,
                  'bg-surface-default': !props.disabled,
                }
              )
            } />) : null
          }
        </div>
      </div>
      {props.children}
    </label>
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
