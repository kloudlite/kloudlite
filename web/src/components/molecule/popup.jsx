import React, {Fragment, useRef} from 'react';
import { Dialog, Transition } from '@headlessui/react'

import PropTypes from 'prop-types';
import classnames from "classnames";
import { BounceIt } from "../bounce-it.jsx";
import { XFill } from '@jengaicons/react';
import { Button } from '../atoms/button.jsx';

export const Popup = ({ header, body, show, onClose, secondaryAction, action }) => {
    const closeButtonRef = useRef();
    const closeButtonProps = useButton({ onPress: onClose }, closeButtonRef);
    return (
        <>
            <Transition appear show={show} as={Fragment}>
                <Dialog as="div" className="relative z-10" onClose={onClose}>
                    <Transition.Child
                        as={Fragment}
                        enter="ease-out duration-300"
                        enterFrom="opacity-0"
                        enterTo="opacity-100"
                        leave="ease-in duration-200"
                        leaveFrom="opacity-100"
                        leaveTo="opacity-0"
                    >
                        <div className="fixed inset-0 bg-black bg-opacity-25" />
                    </Transition.Child>

                    <div className="fixed inset-0 overflow-y-auto">
                        <div className="flex min-h-full items-center justify-center p-4">
                            <Transition.Child
                                as={Fragment}
                                enter="ease-out duration-300"
                                enterFrom="opacity-0 scale-95"
                                enterTo="opacity-100 scale-100"
                                leave="ease-in duration-200"
                                leaveFrom="opacity-100 scale-100"
                                leaveTo="opacity-0 scale-95"
                            >
                                <Dialog.Panel className="w-full max-w-md transform overflow-hidden rounded bg-surface-default shadow-modal border border-border-default transition-all">
                                    <Dialog.Title
                                        as="h3"
                                    >
                                        <div className='flex flex-row p-5 border-b border-border-default items-center'>
                                            <span className='headingMd flex-1'>{header}</span>
                                            <BounceIt>
                                                <span {...closeButtonProps} ref={closeButtonRef}
                                                    className={classnames("outline-none flex items-center rounded ring-offset-1 focus:ring-2 focus:ring-border-focus justify-center")} onPress={onClose}>
                                                    <XFill size={24} color="currentColor" />
                                                </span>
                                            </BounceIt>
                                        </div>
                                    </Dialog.Title>
                                    <div className='p-5 bodyMd'>
                                        {body}
                                    </div>
                                    {(action || secondaryAction) && <div className='flex flex-row gap-2 p-5 justify-end'>
                                        {secondaryAction && <Button label={secondaryAction.content} style='outline' onClick={secondaryAction.onAction} />}
                                        {action && <Button label={action.content} onClick={action.onAction} />}
                                    </div>}
                                </Dialog.Panel>
                            </Transition.Child>
                        </div>
                    </div>
                </Dialog>
            </Transition>
        </>
    )
}


Popup.propTypes = {
    header: PropTypes.string.isRequired,
    body: PropTypes.any.isRequired,
    onClose: PropTypes.func,
    action: PropTypes.shape({
        content: PropTypes.string.isRequired,
        onAction: PropTypes.func
    }),
    secondaryAction: PropTypes.shape({
        content: PropTypes.string.isRequired,
        onAction: PropTypes.func
    }),
    show: PropTypes.bool.isRequired
};

Popup.defaultProps = {
    header: "Heading",
    body: <p>
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
    show: true
};
