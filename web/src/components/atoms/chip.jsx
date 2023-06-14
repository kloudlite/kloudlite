import React, { useRef } from 'react';
import PropTypes from 'prop-types';
import classnames from "classnames";
import { XFill } from '@jengaicons/react';

export const Chip = ({ label, disabled, showClose, onClose }) => {

    return (
        <div
            className={classnames(
                "rounded border bodyMd-medium px-2 py-0.5 flex items-center transition-all outline-none",
                "ring-offset-1 focus-visible:ring-2 focus:ring-border-focus w-fit",
                {
                    "text-text-default": !disabled,
                    "text-text-disabled": disabled,
                }, {
                "pointer-events-none": disabled,
            }, {
                "border-border-default": !disabled,
                "border-border-disabled": disabled,
            }, {
                "bg-surface-default": !disabled,
            })}>

            <span className='flex items-center mr-1.5'>
                {label}
            </span>
            {
                showClose && <button
                    disabled={disabled}
                    onClick={onClose}
                    className={classnames('outline-none flex items-center rounded ring-offset-1 focus-visible:ring-2 focus:ring-border-focus justify-center',
                        {
                            "cursor-default": disabled
                        })}>
                    <XFill size={20} color="currentColor" />
                </button>
            }
        </div>
    );
};

Chip.propTypes = {
    label: PropTypes.string.isRequired,
    disabled: PropTypes.bool,
    showClose: PropTypes.bool,
    onClose: PropTypes.func,
};

Chip.defaultProps = {
    label: "test",
    onClose: null,
    disabled: false,
    showClose: true,
};
