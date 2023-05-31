import React, { cloneElement, useRef, useState } from 'react';
import PropTypes from 'prop-types';
import classnames from "classnames";
import { BounceIt } from "../bounce-it.jsx";
import { ArrowsDownUp, Check, XFill } from '@jengaicons/react';
import { Pressable } from '@ark-ui/react';
import { LayoutGroup, motion } from 'framer-motion';


export const ActionButton = ({ label, disabled, critical, active, onClick, LeftIconComp, RightIconComp }) => {
    return (
        <div className="flex flex-row gap-x-1 ring-offset-1 focus-visible:ring-2 focus:ring-border-focus">
            {
                active && <motion.div layoutId='line' className='w-[3px] bg-icon-primary rounded'></motion.div>
            }
            {
                !active && <motion.div layoutId='line_1' className='w-[3px] bg-transparent rounded'></motion.div>
            }
            <BounceIt>
                <Pressable
                    className={classnames(
                        "rounded border bodyMd px-3 py-1 flex gap-1 items-center items-center cursor-pointer transition-all outline-none w-fit border-none px-4 py-2",
                        {
                            "text-text-primary": active,
                            "text-text-disabled": disabled,
                            "text-text-danger hover:text-text-on-primary hover:text-text-on-primary": critical
                        },
                        {
                            "pointer-events-none": disabled,
                        },
                        {
                            "bg-none hover:bg-surface-hovered active:bg-surface-pressed": !active && !disabled && !critical,
                            "bg-none hover:bg-surface-danger-hovered active:bg-surface-danger-pressed": !active && !disabled && critical,
                            "bg-none": disabled,
                            "bg-surface-primary-selected": !critical && active,
                            // "bg-surface-danger-selected": critical,
                        })} onPress={onClick}>
                    {
                        LeftIconComp && <LeftIconComp size={16} color="currentColor" />
                    }
                    {label}
                    {
                        RightIconComp && <RightIconComp size={16} color="currentColor" />
                    }
                </Pressable>
            </BounceIt>

        </div>
    )
}


export const ActionList = ({ children }) => {
    const [active, setActive] = useState(0)
    return (
        <div className='flex flex-col gap-y-3'>
            <LayoutGroup>
                {children.map((child, index) => {
                    return cloneElement(child, {
                        onClick: () => {
                            setActive(index)
                        },
                        active: active == index,
                        key: index
                    })
                })}
            </LayoutGroup>
        </div>
    )
}

// ActionButton.propTypes = {
//     label: PropTypes.string.isRequired,
//     active: PropTypes.bool,
//     onClick: PropTypes.func,
//     disabled: PropTypes.bool,
// };

// ActionButton.defaultProps = {
//     label: "test",
//     active: false,
//     onClick: null,
//     disabled: false,
// };
