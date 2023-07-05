import React, { cloneElement, useState } from 'react';
import * as TG from '@radix-ui/react-toggle-group';
import { ButtonBase } from './button';
import classNames from 'classnames';


const Button = ({ content, value, sharpLeft, sharpRight, className, selected }) => {
    return <TG.Item value={value} aria-label="Center aligned" asChild>
        <ButtonBase variant={'basic'} content={content} sharpLeft={sharpLeft} sharpRight={sharpRight} className={className} selected={selected} />
    </TG.Item>
}

const IconButton = ({ value, sharpLeft, sharpRight, className, icon }) => {
    return <TG.Item value={value} aria-label="Center aligned" asChild>
        <ButtonBase variant={'basic'} iconOnly={true} sharpLeft={sharpLeft} sharpRight={sharpRight} className={className} iconComp={icon} />
    </TG.Item>
}

const ToggleGroup = ({ children, value = "" }) => {
    const [v, setV] = useState(value)

    console.log(children);
    return <TG.Root
        className="bg-surface-default rounded shadow-button flex flex-row"
        type="single"
        value={v}
        onValueChange={(value) => {
            if (value) setV(value);
        }}
        aria-label="Text alignment"
    >
        {Array.isArray(children) ? children.map((child, index) => cloneElement(child, {
            sharpRight: index < children.length - 1,
            sharpLeft: index > 0,
            className: classNames({ "-ml-px": ((index > 0) || (index < children.length - 1)) }),
            selected: (child.props.value == v),
            key: "toggle-item-" + index
        })) : children}
    </TG.Root>
};

ToggleGroup.Button = Button
ToggleGroup.IconButton = IconButton

export default ToggleGroup;