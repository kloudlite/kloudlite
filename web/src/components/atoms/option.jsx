import classnames from 'classnames';
import React, { cloneElement, forwardRef, useEffect, useId, useRef, useState } from 'react';
import { motion } from "framer-motion";
import { AnimatePresence } from 'framer-motion';
import { useTreeState, useMenuTriggerState, Section } from 'react-stately';
import { useMenuItem, useMenu, usePopover, DismissButton, Overlay, useMenuSection, useSeparator, useFocusRing } from 'react-aria';
import { useMenuTrigger } from 'react-aria';
import { Button, ButtonBase } from './button';
import { TextInput } from './input';
import { Item } from 'react-stately';


export const OptionItemTypes = Object.freeze({
    RADIO: "RADIO",
    LABEL: "LABEL",
    CHECKBOX: "CHECKBOX"
})

const Popover = (props) => {
    let ref = useRef(null);
    let { state, children } = props;

    let { popoverProps, underlayProps } = usePopover(
        {
            ...props,
            popoverRef: ref
        },
        state
    );

    return (
        <AnimatePresence>
            {state.isOpen && <Overlay>
                <div {...underlayProps} className="fixed inset-0" />
                <motion.div
                    {...popoverProps}
                    ref={ref}
                    key={"optionlist"}
                    className="z-10 border border-border-default shadow-popover bg-surface-default rounded rounded-tr-none"
                    initial={{ opacity: 0, translateY: 1, scale: 0.99 }}
                    animate={{ opacity: 1, translateY: 0, scale: 1 }}
                    exit={{ opacity: 0, translateY: 1, scale: 0.99 }}
                    transition={{
                        duration: 0.2,
                        ease: 'anticipate',
                    }}
                >
                    <DismissButton onDismiss={state.close} />
                    {children}
                    <DismissButton onDismiss={state.close} />
                </motion.div>
            </Overlay>}
        </AnimatePresence>
    );
}

function MenuItem({ item, state }) {
    // Get props for the menu item element
    let ref = useRef(null);
    let { menuItemProps, isFocused, isSelected } = useMenuItem(
        { key: item.key, closeOnSelect: false },
        state,
        ref
    );


    let Icon = null
    if (item?.value?.icon) {
        Icon = item.value.icon
    }

    let type = item?.value?.type

    if (type != "search" && !Object.values(OptionItemTypes).includes(type)) {
        type = OptionItemTypes.LABEL
    }


    let { isFocusVisible, focusProps } = useFocusRing();

    let searchRef = useRef(null)

    const setSearchFocus = () => {
        searchRef.current.focus()
    }

    if (type == 'search') {
        menuItemProps.onClick = setSearchFocus
        menuItemProps.onKeyDown = setSearchFocus
        menuItemProps.onMouseDown = setSearchFocus
        menuItemProps.onPointerDown = setSearchFocus
        menuItemProps.onPointerUp = setSearchFocus
    }



    return (
        <li
            {...menuItemProps}
            {...focusProps}
            onPointerEnter={() => { }}
            ref={ref}
            className={classnames('relative outline-none py-2 px-3 cursor-pointer flex flex-row gap-2.5 items-center bodyMd', {
                "focus-visible:ring-2 focus:ring-border-focus z-20": isFocusVisible,
                "bg-surface-primary-selected text-text-primary": isSelected && (type === OptionItemTypes.LABEL),
                "hover:bg-surface-hovered": !isSelected && type != "search",
                "active:bg-surface-pressed": type != "search"
            })}
        >
            {
                type === OptionItemTypes.RADIO && <div className={classnames(
                    "w-5 h-5 rounded-full border ring-border-focus ring-offset-1 transition-all flex items-center justify-center border-border-default",
                    {
                        "border-border-default": !isSelected,
                        "border-border-primary": isSelected,
                    }
                )}>
                    {isSelected && (<div className={classnames(
                        "block w-3 h-3 rounded-full bg-surface-primary-default",
                    )}></div>)}
                </div>
            }

            {
                type == OptionItemTypes.CHECKBOX && <div className={classnames("rounded flex flex-row items-center justify-center border w-5 h-5 outline-none",
                    {
                        "border-border-default": !isSelected,
                        "border-border-primary bg-surface-primary-default": isSelected,
                    }
                )}>
                    <svg width="17" height="16" viewBox="0 0 17 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                        {
                            isSelected && <path d="M14.5 4.00019L6.5 11.9998L2.5 8.00019" className={
                                classnames("stroke-text-on-primary")
                            } strokeLinecap="round" strokeLinejoin="round" />
                        }
                    </svg>

                </div>
            }
            {
                type == OptionItemTypes.LABEL && Icon && <Icon size={16} color="currentColor" />
            }
            {
                type == "search" && <TextInput ref={searchRef} placeholder={"Filter"} />
            }
            {item.rendered}
        </li>
    );
}

function MenuSection({
    section,
    state,
    type
}) {
    let { itemProps, groupProps } = useMenuSection({
        heading: section.rendered,
        "aria-label": section["aria-label"]
    });

    let { separatorProps } = useSeparator({
        elementType: "li"
    });

    console.log(section, state);
    return (
        <>
            <li {...itemProps}>
                <ul {...groupProps}>
                    {[...section.childNodes].map((node) => (
                        <MenuItem
                            key={node.key}
                            item={node}
                            state={state}
                            type={type}
                        />
                    ))}
                    {!section.value.isLast && < li
                        {...separatorProps}
                        className="border-t border-border-disabled my-1"
                    />}
                </ul>
            </li>
        </>
    );
}

const MenuBase = (props) => {
    // Create menu state based on the incoming props
    let state = useTreeState({
        ...props,
    });

    // Get props for the menu element
    let ref = useRef(null);
    let { menuProps } = useMenu({
        ...props,
    }, state, ref);


    return (
        <div>
            <ul
                {...menuProps}
                ref={ref}
                style={{

                    width: 192
                }}
                className={classnames('outline-none m-0 p-0 py-2 list-none',)}
            >
                {[...state.collection].map((item) => (
                    <MenuSection
                        key={item.key}
                        section={item}
                        state={state}
                        type={props.type}
                    />
                ))}
            </ul>
        </div>
    );
}


const OptionListBase = (props) => {

    let state = useMenuTriggerState({ ...props });

    // Get props for the button and menu elements
    let ref = useRef(null);
    let { menuTriggerProps, menuProps } = useMenuTrigger({}, state, ref);
    return (
        <div>
            {cloneElement(props.trigger,
                {
                    ...menuTriggerProps,
                    ref: ref,
                    sharpLeft: props.sharpLeft,
                    sharpRight: props.sharpRight,
                    className: classnames({ "-ml-px": (props.sharpLeft || props.sharpRight) }),
                    size: props.size,
                    style: props.style,
                    selected: state.isOpen
                })}
            {state.isOpen &&
                (
                    <Popover state={state} triggerRef={ref} placement="bottom end">

                        <MenuBase
                            {...props}
                            {...menuProps}
                        />
                    </Popover>
                )}
        </div>
    );
}

export const OptionList = ({ items, size, style = "basic" }) => {
    return <div className={classnames("flex flex-row")}>

        {items && items.map((child, index) => {

            const sharpRight = index < items.length - 1;
            const sharpLeft = index > 0;

            let childItems = child.items.map((item, index) => ({ id: item.id, children: item.children, isLast: index === child.items.length - 1, type: null }))
            if (childItems.searchFilter) {
                childItems = [{ id: useId(), type: "search", isLast: true, children: [{ id: 0, type: "search" }] }, ...child.items]
            }
            return <OptionListBase
                key={child.key}
                style={style}
                size={size}
                sharpLeft={sharpLeft}
                sharpRight={sharpRight}
                items={childItems}
                selectionMode={'single'}
                selectedKeys={child.selected}
                onSelectionChange={child?.onSelectionChange}
                searchFilter={child?.searchFilter}
                onSearchFilterChange={child?.onSearchFilterChange}
                trigger={child.trigger}
            >
                {item => (
                    <Section items={item.children} key={item.id}>
                        {item => <Item key={item.id}>{item.label}</Item>}
                    </Section>
                )}
            </OptionListBase>
        })
        }
    </div>
}