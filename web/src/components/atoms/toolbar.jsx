import * as React from 'react';
import { composeEventHandlers } from '@radix-ui/primitive';
import { createContextScope } from '@radix-ui/react-context';
import * as RovingFocusGroup from '@radix-ui/react-roving-focus';
import { createRovingFocusGroupScope } from '@radix-ui/react-roving-focus';
import { Primitive } from '@radix-ui/react-primitive';
import * as SeparatorPrimitive from '@radix-ui/react-separator';
import * as ToggleGroupPrimitive from '@radix-ui/react-toggle-group';
import { createToggleGroupScope } from '@radix-ui/react-toggle-group';
import { useDirection } from '@radix-ui/react-direction';

/* -------------------------------------------------------------------------------------------------
 * Toolbar
 * -----------------------------------------------------------------------------------------------*/

const TOOLBAR_NAME = 'Toolbar';

const [createToolbarContext, createToolbarScope] = createContextScope(TOOLBAR_NAME, [
    createRovingFocusGroupScope,
    createToggleGroupScope,
]);
const useRovingFocusGroupScope = createRovingFocusGroupScope();
const useToggleGroupScope = createToggleGroupScope();


const [ToolbarProvider, useToolbarContext] =
    createToolbarContext < ToolbarContextValue > (TOOLBAR_NAME);


const Toolbar = React.forwardRef(
    (props, forwardedRef) => {
        const { __scopeToolbar, orientation = 'horizontal', dir, loop = true, ...toolbarProps } = props;
        const rovingFocusGroupScope = useRovingFocusGroupScope(__scopeToolbar);
        const direction = useDirection(dir);
        return (
            <ToolbarProvider scope={__scopeToolbar} orientation={orientation} dir={direction}>
                <RovingFocusGroup.Root
                    asChild
                    {...rovingFocusGroupScope}
                    orientation={orientation}
                    dir={direction}
                    loop={loop}
                >
                    <Primitive.div
                        role="toolbar"
                        aria-orientation={orientation}
                        dir={direction}
                        {...toolbarProps}
                        ref={forwardedRef}
                    />
                </RovingFocusGroup.Root>
            </ToolbarProvider>
        );
    }
);

Toolbar.displayName = TOOLBAR_NAME;

/* -------------------------------------------------------------------------------------------------
 * ToolbarSeparator
 * -----------------------------------------------------------------------------------------------*/

const SEPARATOR_NAME = 'ToolbarSeparator';


const ToolbarSeparator = React.forwardRef(
    (props, forwardedRef) => {
        const { __scopeToolbar, ...separatorProps } = props;
        const context = useToolbarContext(SEPARATOR_NAME, __scopeToolbar);
        return (
            <SeparatorPrimitive.Root
                orientation={context.orientation === 'horizontal' ? 'vertical' : 'horizontal'}
                {...separatorProps}
                ref={forwardedRef}
            />
        );
    }
);

ToolbarSeparator.displayName = SEPARATOR_NAME;

/* -------------------------------------------------------------------------------------------------
 * ToolbarButton
 * -----------------------------------------------------------------------------------------------*/

const BUTTON_NAME = 'ToolbarButton';

const ToolbarButton = React.forwardRef(
    (props, forwardedRef) => {
        const { __scopeToolbar, ...buttonProps } = props;
        const rovingFocusGroupScope = useRovingFocusGroupScope(__scopeToolbar);
        return (
            <RovingFocusGroup.Item asChild {...rovingFocusGroupScope} focusable={!props.disabled}>
                <Primitive.button type="button" {...buttonProps} ref={forwardedRef} />
            </RovingFocusGroup.Item>
        );
    }
);

ToolbarButton.displayName = BUTTON_NAME;

/* -------------------------------------------------------------------------------------------------
 * ToolbarLink
 * -----------------------------------------------------------------------------------------------*/

const LINK_NAME = 'ToolbarLink';

const ToolbarLink = React.forwardRef(
    (props, forwardedRef) => {
        const { __scopeToolbar, ...linkProps } = props;
        const rovingFocusGroupScope = useRovingFocusGroupScope(__scopeToolbar);
        return (
            <RovingFocusGroup.Item asChild {...rovingFocusGroupScope} focusable>
                <Primitive.a
                    {...linkProps}
                    ref={forwardedRef}
                    onKeyDown={composeEventHandlers(props.onKeyDown, (event) => {
                        if (event.key === ' ') event.currentTarget.click();
                    })}
                />
            </RovingFocusGroup.Item>
        );
    }
);

ToolbarLink.displayName = LINK_NAME;

/* -------------------------------------------------------------------------------------------------
 * ToolbarToggleGroup
 * -----------------------------------------------------------------------------------------------*/

const TOGGLE_GROUP_NAME = 'ToolbarToggleGroup';

const ToolbarToggleGroup = React.forwardRef(
    (
        props,
        forwardedRef
    ) => {
        const { __scopeToolbar, ...toggleGroupProps } = props;
        const context = useToolbarContext(TOGGLE_GROUP_NAME, __scopeToolbar);
        const toggleGroupScope = useToggleGroupScope(__scopeToolbar);
        return (
            <ToggleGroupPrimitive.Root
                data-orientation={context.orientation}
                dir={context.dir}
                {...toggleGroupScope}
                {...toggleGroupProps}
                ref={forwardedRef}
                rovingFocus={false}
            />
        );
    }
);

ToolbarToggleGroup.displayName = TOGGLE_GROUP_NAME;

/* -------------------------------------------------------------------------------------------------
 * ToolbarToggleItem
 * -----------------------------------------------------------------------------------------------*/

const TOGGLE_ITEM_NAME = 'ToolbarToggleItem';

const ToolbarToggleItem = React.forwardRef(
    (props, forwardedRef) => {
        const { __scopeToolbar, ...toggleItemProps } = props;
        const toggleGroupScope = useToggleGroupScope(__scopeToolbar);
        const scope = { __scopeToolbar: props.__scopeToolbar };

        return (
            <ToolbarButton asChild {...scope}>
                <ToggleGroupPrimitive.Item {...toggleGroupScope} {...toggleItemProps} ref={forwardedRef} />
            </ToolbarButton>
        );
    }
);

ToolbarToggleItem.displayName = TOGGLE_ITEM_NAME;

export const Toolbar = Toolbar

export const ToolbarSeparator
export const ToolbarButton,
export const ToolbarLink,
export const ToolbarToggleGroup,
export const ToolbarToggleItem,
//
export const ToolbarRoot,
export const ToolbarSeparator,
export const ToolbarButton,
export const ToolbarLink,
export const ToolbarToggleGroup,
export const ToolbarToggleItem