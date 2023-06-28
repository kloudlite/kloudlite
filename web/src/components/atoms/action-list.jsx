import React, { useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import classnames from "classnames";
import { LayoutGroup, motion } from 'framer-motion';
import { Link } from "@remix-run/react"

export const ActionButton = ({
    label,
    disabled,
    critical,
    active,
    onClick,
    href,
    LeftIconComp,
    RightIconComp,
    rightEmptyPlaceholder
}) => {

    return (
        <div className={classnames("w-full flex flex-row gap-x-1")}>
            {
                active && <motion.div layoutId='line' className='w-0.5 bg-icon-primary rounded'></motion.div>
            }
            {
                !active && <motion.div layoutId='line_1' className='w-0.5 bg-transparent rounded'></motion.div>
            }
            <Link
                to={href}
                className={classnames(
                    "w-[inherit] rounded border bodyMd flex gap-1 items-center justify-between cursor-pointer outline-none border-none px-4 py-2 ring-offset-1 focus-visible:ring-2 focus:ring-border-focus",
                    {
                        "text-text-primary bodyMd-medium": active,
                        "text-text-disabled": disabled,
                        "text-text-danger hover:text-text-on-primary active:text-text-on-primary": critical,
                    },
                    {
                        "pointer-events-none": disabled,
                    },
                    {
                        "bg-none hover:bg-surface-hovered active:bg-surface-pressed": !active && !disabled && !critical,
                        "bg-none hover:bg-surface-danger-hovered active:bg-surface-danger-pressed": !active && !disabled && critical,
                        "bg-none": disabled,
                        "bg-surface-primary-selected": !critical && active,
                    })} onClick={!critical ? onClick : null}>
                <div className='flex flex-row items-center gap-1'>
                    {
                        LeftIconComp && <LeftIconComp size={16} color="currentColor" />
                    }
                    {label}
                </div>
                {
                    RightIconComp && <RightIconComp size={16} color="currentColor" />
                }
                {
                    !RightIconComp && rightEmptyPlaceholder && <div className='w-4 h-4'></div>
                }
            </Link>

        </div>
    )
}


export const ActionList = ({ items, value, onChange, layoutId }) => {
    const [active, setActive] = useState(value)
    useEffect(() => {
        if (onChange) onChange(active)
    }, [active])
    return (
        <div className={classnames('flex flex-col gap-y-1')}>
            <LayoutGroup id={layoutId}>
                {items.map((child, index) => {
                    return <ActionButton
                        critical={child.critical}
                        label={child.label}
                        href={child.href}
                        LeftIconComp={child.LeftIconComp}
                        RightIconComp={child.RightIconComp}
                        rightEmptyPlaceholder={!child.RightIconComp}
                        key={child.key}
                        active={child.value === active}
                        onClick={() => {
                            setActive(child.value);
                        }}
                    />
                })}
            </LayoutGroup>
        </div>
    )
}

ActionButton.propTypes = {
    label: PropTypes.string.isRequired,
    href: PropTypes.string.isRequired,
    active: PropTypes.bool,
    onClick: PropTypes.func,
    disabled: PropTypes.bool,
};

ActionButton.defaultProps = {
    label: "test",
    href: "#",
    active: false,
    onClick: null,
    disabled: false,
};

ActionList.propTypes = {
    items: PropTypes.arrayOf(PropTypes.shape({
        label: PropTypes.string.isRequired,
        value: PropTypes.oneOfType([PropTypes.string, PropTypes.object]).isRequired,
        key: PropTypes.string,
        RightIconComp: PropTypes.object,
        LeftIconComp: PropTypes.object,
    })).isRequired,
    value: PropTypes.oneOfType([PropTypes.string, PropTypes.object]).isRequired,
    onChange: PropTypes.func,
    layoutId: PropTypes.string.isRequired
}
