import React, {useRef} from 'react';
import PropTypes from 'prop-types';
import classnames from "classnames";
import {BounceIt} from "../bounce-it.jsx";
import {XFill} from '@jengaicons/react';
import {useButton} from "react-aria";


/**
 * Button component for user interaction
 */
export const Chip = ({label, disabled, selected, showClose, onClose, LeftIconComp}) => {
    const closeButtonRef = useRef();
    const closeButtonProps = useButton({onPress: onClose}, closeButtonRef);
    return (
        <div
            className={classnames(
                "rounded-full border bodyMd-medium px-3 py-1 flex gap-1 items-center cursor-pointer transition-all outline-none",
                "ring-offset-1 focus-visible:ring-2 focus:ring-border-focus w-fit",
                {
                    "text-text-default": !selected && !disabled,
                    "text-text-disabled": disabled,
                    "text-text-on-primary": selected,
                }, {
                    "pointer-events-none": disabled,
                }, {
                    "border-border-default": !selected && !disabled,
                    "border-border-disabled": disabled,
                    "border-border-primary": selected,
                }, {
                    "bg-grey-50 hover:bg-zinc-200 active:bg-zinc-300": !selected && !disabled,
                    "bg-grey-50": disabled,
                    "bg-primary-700": selected,
                })}>

      <span className='flex items-center gap-1'>
      {
          LeftIconComp && <LeftIconComp size={16} color="currentColor"/>
      }
          {label}
      </span>
            {
                showClose &&
                <BounceIt>
            <span ref={closeButtonRef} {...closeButtonProps} className='flex items-center'>
              <XFill size={16} color="currentColor"/>
            </span>
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
    onClose: PropTypes.func,
};

Chip.defaultProps = {
    label: "test",
    selected: false,
    onClick: null,
    onClose: null,
    disabled: false,
    showClose: true,
};
