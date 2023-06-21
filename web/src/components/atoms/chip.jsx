import React, { useRef } from 'react';
import PropTypes from 'prop-types';
import classnames from "classnames";
import { MagicWandFill, XFill } from '@jengaicons/react';
import { AriaButton } from './button';
import { useFocusRing } from 'react-aria';


export const ChipTypes = Object.freeze({
    BASIC: 0,
    REMOVABLE: 1,
    CLICKABLE: 2
})




export const Chip = ({ label, disabled, type = ChipTypes.BASIC, onClose, prefix, onClick }) => {
    let { isFocusVisible, focusProps } = useFocusRing();

    let Component = "div"
    if (type === ChipTypes.CLICKABLE)
        Component = AriaButton

    let Prefix = prefix


    return (
        <Component
            {...focusProps}
            onPress={type === ChipTypes.CLICKABLE ? onClick : undefined}
            className={classnames(
                "rounded border bodySm-medium py-px flex items-center transition-all outline-none flex-row gap-1.5 ring-offset-1",
                "w-fit",
                {
                    "text-text-default": !disabled,
                    "text-text-disabled": disabled,
                },
                {
                    "pointer-events-none": disabled,
                },
                {
                    "border-border-default": !disabled,
                    "border-border-disabled": disabled,
                },
                {
                    "bg-surface-default": !disabled,
                },
                {
                    "pr-1 pl-2": type === ChipTypes.REMOVABLE,
                    "px-2": type != ChipTypes.REMOVABLE
                },
                {
                    "hover:bg-surface-hovered active:bg-surface-pressed ": type === ChipTypes.CLICKABLE,
                    "focus-visible:ring-2 focus:ring-border-focus": isFocusVisible && type === ChipTypes.CLICKABLE
                }
            )}>
            {
                Prefix && type != ChipTypes.CLICKABLE && ((typeof Prefix == "string") ? <span className='bodySm text-text-soft'>{Prefix}</span> : <Prefix size={16} color="currentColor" />)
            }
            <span className='flex items-center'>
                {label}
            </span>
            {
                type == ChipTypes.REMOVABLE && <AriaButton
                    disabled={disabled}
                    onPress={onClose}
                    {...focusProps}
                    className={classnames('outline-none flex items-center rounded-sm ring-offset-0 justify-center hover:bg-surface-hovered active:bg-surface-pressed',
                        {
                            "cursor-default": disabled
                        },
                        {
                            "focus-visible:ring-2 focus:ring-border-focus": isFocusVisible
                        })}>
                    <XFill size={16} color="currentColor" />
                </AriaButton>
            }
        </Component>
    );
};




Chip.propTypes = {
    label: PropTypes.string.isRequired,
    disabled: PropTypes.bool,
    onClose: PropTypes.func,
};

Chip.defaultProps = {
    label: "test",
    onClose: null,
    disabled: false,
};
