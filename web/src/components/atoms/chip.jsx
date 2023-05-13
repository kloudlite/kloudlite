import React, { useRef } from 'react';
import PropTypes from 'prop-types';
import classnames from "classnames";
import {BounceIt} from "../bounce-it.jsx";
import { useButton } from 'react-aria';
import { X, XFill } from '@jengaicons/react';

/**
 * Button component for user interaction
 */
export const Chip = ({label, disabled, selected, showClose, onClick,onClose, LeftIconComp}) => {
  const closeButtonRef = useRef();
  const buttonRef = useRef();
  const {buttonProps:closeButtonProps} = useButton({onPress:onClose}, closeButtonRef);
  const {buttonProps} = useButton({onPress:onClick}, buttonRef);
  return (
      <div 
        className={classnames(
        "rounded-full border bodyMd-medium px-3 py-1 flex gap-1 items-center cursor-pointer transition-all outline-none",
        "ring-offset-1 focus-visible:ring-2 focus:ring-border-focus w-fit",
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
        })}>
        <BounceIt>
          <button className='flex items-center gap-1' {...buttonProps} ref={buttonRef}>
          {
            LeftIconComp && <LeftIconComp size={16} color="currentColor" />
          }
          {label}
          </button>
        </BounceIt>
        {
          showClose && 
          <BounceIt>
            <button {...closeButtonProps} ref={closeButtonRef} className='flex items-center'>
              <XFill size={16} color="currentColor" />
            </button>
          </BounceIt>
        }
      </div>
  );
};

Chip.propTypes = {
  label: PropTypes.string.isRequired,
  selected: PropTypes.bool,
  onClick: PropTypes.func,
  disabled: PropTypes.bool,
  showClose: PropTypes.bool,
  onClose:PropTypes.func,
};

Chip.defaultProps = {
  label: "test",
  selected: false,
  onClick: null,
  onClose:null,
  disabled: false,
  showClose: true,
};
