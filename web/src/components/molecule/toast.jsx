import classNames from "classnames"
import PropTypes from "prop-types"
import { useRef, useState } from "react";
import { Button } from "../atoms/button";
import * as ToastRadix from '@radix-ui/react-toast';

export const ToastProvider = ({ children, duration }) => {

    return (
        <ToastRadix.Provider swipeDirection="right" duration={duration}>
            {children}
            <ToastRadix.Viewport className="flex flex-col gap-2 fixed bottom-2 right-2 m-0 list-none z-[2147483647] outline-none" />
        </ToastRadix.Provider>
    );
}

export const Toast = ({ show }) => {
    const [open, setOpen] = useState(show)
    return (
        <ToastRadix.Root
            className={classNames("toast flex flex-row bg-surface-tertiary-default border-border-tertiary rounded shadow-popover text-text-on-primary p-3 gap-3 items-center justify-between",
                "data-[state=open]:animate-slideIn data-[state=closed]:animate-slideOut data-[swipe=move]:translate-x-[var(--radix-toast-swipe-move-x)] data-[swipe=cancel]:translate-x-0 data-[swipe=cancel]:transition-[transform_200ms_ease-out] data-[swipe=end]:animate-swipeOut")}
            open={open}
            onOpenChange={setOpen}
        >
            <ToastRadix.Title className="bodyMd-medium">
                Scheduled: Catch up
            </ToastRadix.Title>
            <ToastRadix.Action asChild altText="undo">
                <Button content={"Undo"} variant="secondary-plain" size="small" className={"text-text-surface-primary"} />
            </ToastRadix.Action>
        </ToastRadix.Root>

    );
}