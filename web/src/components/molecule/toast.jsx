import classNames from "classnames"
import PropTypes from "prop-types"
import { useRef } from "react";
import { useToast, useToastRegion } from "@react-aria/toast";
import { useToastState } from "@react-stately/toast";
import { Button, IconButton } from "../atoms/button";
import { XFill } from "@jengaicons/react";

export const ToastProvider = ({ children, ...props }) => {
    let state = useToastState({
        maxVisibleToasts: 5,
        hasExitAnimation: true
    });

    return (
        <>
            {children(state)}
            {state.visibleToasts.length > 0 && (
                <ToastRegion {...props} state={state} />
            )}
        </>
    );
}

function ToastRegion({ state, ...props }) {
    let ref = useRef(null);
    let { regionProps } = useToastRegion(props, state, ref);

    return (
        <div {...regionProps} ref={ref} className="flex flex-col gap-2 fixed bottom-4 right-4">
            {state.visibleToasts.map(toast => (
                <Toast key={toast.key} toast={toast} state={state} />
            ))}
        </div>
    );
}

function Toast({ state, ...props }) {
    let ref = useRef(null);
    let { toastProps, titleProps, closeButtonProps } = useToast(props, state, ref);

    console.log(props);
    return (
        <div {...toastProps} ref={ref} className="toast flex flex-row bg-surface-tertiary-default border-border-tertiary rounded shadow-popover text-text-on-primary p-3 gap-5"
            // Use a data attribute to trigger animations in CSS.
            data-animation={props.toast.animation}
            onAnimationEnd={() => {
                // Remove the toast when the exiting animation completes.
                if (props.toast.animation === 'exiting') {
                    state.remove(props.toast.key);
                }
            }}
        >
            <div {...titleProps} className="bodyMd-medium">{props.toast.content}</div>
            <Button label={"Undo"} style="plain" size="small" />
        </div>
    );
}