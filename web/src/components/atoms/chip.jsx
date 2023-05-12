import React, { useRef } from 'react';
import PropTypes from 'prop-types';
import classnames from "classnames";
import {BounceIt} from "../bounce-it.jsx";
import { useButton } from 'react-aria';



/**
 * Button component for user interaction
 */
export const Chip = ({label, disabled, selected, onClick, leftIcon, rightIcon}) => {
  const ref = useRef()
  const {buttonProps} = useButton({onPress:onClick}, ref);
  return (
    <BounceIt>
      <button ref={ref} {...buttonProps} className={classnames(
        "rounded-full border bodyMd-medium flex gap-0.5 px-3 py-1 cursor-pointer transition-all outline-none",
        "ring-offset-1 focus-visible:ring-2 focus:ring-border-focus",
        {
          "text-text-default":!selected && !disabled,
          "text-text-disabled":disabled,
          "text-text-on-primary":selected,
        }, {
          "pointer-events-none":disabled,
        }, {
          "border-border-default": !selected && !disabled,
          "border-border-disabled":disabled,
          "border-border-primary":selected,
        },{
          "bg-grey-50 hover:bg-zinc-200 active:bg-zinc-300": !selected && !disabled,
          "bg-grey-50":disabled,
          "bg-primary-700":selected,
        })} onClick={()=> {
        setSelectedState(()=> {
          return !selectedState
        })
      }}>
        {leftIcon}
        {label}
        {rightIcon}
      </button>
    </BounceIt>
  );
};

Chip.propTypes = {
  label: PropTypes.string.isRequired,
  selected: PropTypes.bool,
  onClick: PropTypes.func,
  disabled: PropTypes.bool,
};

Chip.defaultProps = {
  label: "test",
  selected: false,
  onClick: () => {},
  disabled: false,
};
