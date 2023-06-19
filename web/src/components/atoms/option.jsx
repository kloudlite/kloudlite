import classnames from 'classnames';
import React, { forwardRef, useEffect, useRef, useState, createContext, useContext } from 'react';
import { motion } from "framer-motion";
import { AnimatePresence } from 'framer-motion';
import { useTreeState, useMenuTriggerState, Section } from 'react-stately';
import { useMenuItem, useMenu, usePopover, DismissButton, Overlay, useMenuSection, useSeparator } from 'react-aria';
import { useMenuTrigger } from 'react-aria';
import { useFocusRing, VisuallyHidden, useRadio } from "react-aria";
import { useRadioGroupState } from "react-stately";
import { useRadioGroup } from "react-aria";
import { Button } from './button';
import { TextInput } from './input';
import { Item } from 'react-stately';




let RadioContext = createContext(null);

const RG = (props) => {
    let { items, label, disabled } = props;

    const [value, setValue] = useState(props.value)
    useEffect(() => {
        if (props.onChange) props.onChange(value)
    }, [value])

    let state = useRadioGroupState({
        ...props, value: value, onChange: (e) => {
            setValue(e)
        }, isDisabled: disabled
    });
    let { radioGroupProps, labelProps } =
        useRadioGroup(props, state);
    console.log(items);

    return (
        <div {...radioGroupProps} className="flex flex-col gap-y-2.5">
            <span {...labelProps}>{label}</span>
            <RadioContext.Provider value={state}>
                {items && items.map((item) => {

                    return <Radio label={item.label} disabled={item.disabled} value={item.value} xkey={item.key} xstate={item.state} />
                })}
            </RadioContext.Provider>
        </div>
    );
}

const Radio = ({ label, disabled, value, xkey, xstate }) => {
    let props = { label, disabled, value, xkey, xstate }
    console.log(props);
    let state = useContext(RadioContext);
    let ref = useRef(null);
    let { inputProps, isSelected, isDisabled } = useRadio({ ...props, isDisabled: props.disabled, "aria-label": props.label }, state, ref);
    let { isFocusVisible, focusProps } = useFocusRing();
    let { menuItemProps, isFocused } = useMenuItem(
        { key: props.xkey, closeOnSelect: false },
        props.xstate,
        ref
    );
    return (
        <label
            {...menuItemProps}
            className="flex gap-2 items-center cursor-pointer w-fit group"
            onPointerEnter={() => { }}

        >
            <VisuallyHidden>
                <input {...inputProps} {...focusProps} ref={ref} />
            </VisuallyHidden>
            <div className={classnames("w-5 h-5 rounded-full border group-hover:bg-surface-hovered ring-border-focus ring-offset-1 transition-all flex items-center justify-center",
                isDisabled ? {
                    "border-border-disabled": true,
                } : {
                    "border-border-default": !isSelected,
                    "border-border-primary": isSelected,
                },
                {
                    "ring-2": isFocusVisible
                }
            )}>
                {isSelected && (<div className={classnames(
                    "block w-3 h-3  rounded-full",
                    {
                        "bg-surface-disabled-default": isDisabled,
                        "bg-surface-primary-default": !isDisabled,
                    },
                )}></div>)}
            </div>
            <div className={classnames({
                "text-text-disabled": isDisabled,
                "text-text-default": !isDisabled,
            }, "bodyMd-medium")}>{props.label}</div>
        </label>
    );
}

export const Popover = (props) => {
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
                    className="z-10 border border-border-default shadow-popover bg-surface-default rounded"
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
    let { menuItemProps, isFocused, isSelected, isDisabled } = useMenuItem(
        { key: item.key, closeOnSelect: false },
        state,
        ref
    );


    return (
        <li
            {...menuItemProps}
            onPointerEnter={() => { }}
            ref={ref}
            className={classnames('outline-none py-2 px-3 cursor-pointer first:rounded-t last:rounded-b', {
                "hover:bg-surface-hovered": !isSelected,
                "focus-visible:ring-2 focus:ring-border-focus": isFocused
            })}
        >
            {item.rendered}
        </li>
    );
}

function MenuSection({
    section,
    state,
    onAction,
    onClose
}) {
    let { itemProps, groupProps } = useMenuSection({
        heading: section.rendered,
        "aria-label": section["aria-label"]
    });

    let { separatorProps } = useSeparator({
        elementType: "li"
    });

    // console.log(section.childNodes);
    let Element = section.props.type === "RadioGroup" ? RG : 'div'
    return (
        <>
            {section.key !== state.collection.getFirstKey() && (
                <li
                    {...separatorProps}
                    className="border-t border-border-disabled my-1"
                />
            )}
            <li {...itemProps}>
                {Element === RG && <Element {...groupProps} value={section.value.props.value}
                    optionlist
                    items={[...section.childNodes].map((node) => {
                        console.log(node.key);
                        return { state, label: node.value.props.label, value: node.value.props.value, key: node.value.props.key, }
                    })}
                />}
            </li>
        </>
    );
}

// <MenuItem
//                             key={node.key}
//                             item={node}
//                             state={state}
//                             onAction={onAction}
//                             onClose={onClose}
//                         />
export const MenuBase = (props) => {
    // Create menu state based on the incoming props
    let state = useTreeState(props);

    // Get props for the menu element
    let ref = useRef(null);
    let { menuProps } = useMenu({
        ...props,
    }, state, ref);

    return (
        <div>
            <div className='py-2 px-3 w-[192px]'>
                <TextInput />
            </div>
            <ul
                {...menuProps}
                ref={ref}
                style={{
                    margin: 0,
                    padding: 0,
                    listStyle: 'none',
                    width: 192
                }}
            >
                {[...state.collection].map((item) => (
                    <MenuSection
                        key={item.key}
                        section={item}
                        state={state}
                        onAction={props.onAction}
                        onClose={props.onClose}
                    />
                ))}
            </ul>
        </div>
    );
}


export const OptionListBase = (props) => {
    // Create state based on the incoming props
    let [selected, setSelected] = useState(
        new Set(['copy'])
    );

    const onSelect = (e) => {

    }

    let state = useMenuTriggerState({ ...props });

    // Get props for the button and menu elements
    let ref = useRef(null);
    let { menuTriggerProps, menuProps } = useMenuTrigger({}, state, ref);

    return (
        <>
            <Button
                {...menuTriggerProps}
                ref={ref}
                label={props.label}
            />
            {state.isOpen &&
                (
                    <Popover state={state} triggerRef={ref} placement="bottom end">
                        <MenuBase
                            {...props}
                            {...menuProps}
                            autoFocus={state.focusStrategy || false}
                        />
                    </Popover>
                )}
        </>
    );
}


const OptionList = ({ label, children }) => {
    let [selected, setSelected] = useState(new Set([1, 3]));
    let openWindows = children
    return (
        <OptionListBase
            label={label}
            selectionMode="multiple"
            selectedKeys={selected}
            items={openWindows}
            onSelectionChange={setSelected}>
            {item => (
                <Section items={item.props.children} type={item.type.TagName}>
                    {item => <Item>{1}</Item>}
                </Section>
            )}
        </OptionListBase>
    );
}


const RadioGroup = () => {
}

const RadioGroupItem = () => {

}


OptionList.RadioGroup = RadioGroup
OptionList.RadioGroup.TagName = "RadioGroup"
OptionList.RadioGroup.Item = RadioGroupItem

export default OptionList
