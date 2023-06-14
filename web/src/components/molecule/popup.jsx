import React, { useRef, useState } from 'react';

import PropTypes from 'prop-types';
import classnames from "classnames";
import { XFill } from '@jengaicons/react';
import { Button, IconButton } from '../atoms/button.jsx';
import { useModalOverlay } from 'react-aria';
import { useOverlayTriggerState } from 'react-stately';
import { Overlay } from 'react-aria';
import { useDialog } from 'react-aria';
import { AnimatePresence, motion } from 'framer-motion';


export const Modal = (props) => {
    let ref = useRef(null);
    let { modalProps, underlayProps } = useModalOverlay(props, props.state, ref);

    return (
        <AnimatePresence>
            {props.state.isOpen && <Overlay>
                <motion.div
                    className={classnames('fixed z-10 inset-0 flex items-center justify-center')}
                    style={{ backgroundColor: props.backdrop && 'rgba(0,0,0,0.5)' }}
                    {...underlayProps}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{
                        duration: 0.25,
                        ease: 'anticipate',
                    }}
                    key={"overlay"}
                >
                    <motion.div
                        key={"modal"}
                        {...modalProps}
                        ref={ref}
                        className='w-full max-w-95 md:max-w-153'
                        initial={{ opacity: 0, translateY: 4, scale: 0.97 }}
                        animate={{ opacity: 1, translateY: 0, scale: 1 }}
                        exit={{ opacity: 0, translateY: 4, scale: 0.97 }}
                        transition={{
                            duration: 0.25,
                            ease: 'anticipate',
                        }}
                    >
                        {props.children}
                    </motion.div>
                </motion.div>
            </Overlay>}
        </AnimatePresence>
    )
}


export function AlertDialog(props) {
    let { children, onClose, secondaryAction, action, header } = props;

    let ref = useRef(null);
    let { dialogProps, titleProps } = useDialog(
        {
            ...props,
            role: "alertdialog"
        },
        ref
    );

    return (
        <div {...dialogProps} ref={ref} className="outline-none transform overflow-hidden rounded bg-surface-default shadow-modal border border-border-default transition-all">
            <h3 {...titleProps} className='flex flex-row p-5 border-b border-border-default items-center'>
                <span className='headingMd flex-1'>{header}</span>
                <IconButton IconComp={XFill} style={'plain'} size='small' onClick={onClose} />
            </h3>
            <div className="p-5 bodyMd">{children}</div>
            {(action || secondaryAction) && <div className='flex flex-row gap-2 p-5 justify-end'>
                {secondaryAction && <Button label={secondaryAction.content} style='outline' onClick={secondaryAction.onAction} />}
                {action && <Button label={action.content} onClick={action.onAction} />}
            </div>}
        </div>
    );
}
export const Popup = ({ header, show, onClose, secondaryAction, action, children, backdrop }) => {
    let state = useOverlayTriggerState({ isOpen: show });
    return <Modal state={state} backdrop={backdrop}>
        <AlertDialog header={header} secondaryAction={secondaryAction} action={action} onClose={onClose}>
            {children}
        </AlertDialog>
    </Modal>
}

Popup.propTypes = {
    header: PropTypes.string.isRequired,
    children: PropTypes.any.isRequired,
    onClose: PropTypes.func,
    action: PropTypes.shape({
        content: PropTypes.string.isRequired,
        onAction: PropTypes.func
    }),
    secondaryAction: PropTypes.shape({
        content: PropTypes.string.isRequired,
        onAction: PropTypes.func
    }),
    show: PropTypes.bool.isRequired,
    backdrop: PropTypes.bool
};

Popup.defaultProps = {
    header: "Heading",
    children: <p>
        Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
    </p>,
    action: {
        content: "Ok",
        onAction: () => { }
    },
    secondaryAction: {
        content: "Cancel",
        onAction: () => { }
    },
    onClose: () => { },
    show: true,
    backdrop: false
};
