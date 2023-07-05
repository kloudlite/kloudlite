import PropTypes from 'prop-types';
import { XFill } from '@jengaicons/react';
import { Button } from '../atoms/button.jsx';
import * as Dialog from '@radix-ui/react-dialog';
import { useEffect, useState } from 'react';


export const Popup = ({ header, show, onClose, secondaryAction, action, children, backdrop }) => {
    const [open, setOpen] = useState(show)

    return <Dialog.Root open={open} onOpenChange={setOpen}>
        <Dialog.Portal>
            <Dialog.Overlay className="DialogOverlay bg-blackA9 data-[state=open]:animate-overlayShow fixed inset-0" />
            <Dialog.Content className="DialogContent outline-none transform overflow-hidden rounded bg-surface-default shadow-modal border border-border-default transition-all fixed top-1/2 left-1/2 w-[90vw] max-w-[450px] -translate-x-1/2 -translate-y-1/2">
                <Dialog.Title className="headingMd p-5 border-b border-border-default flex flex-row items-center justify-between">
                    {header}
                    <Dialog.Close asChild>

                        <Button variant={'plain'} size='small' onClick={onClose} content={<XFill size={20} />} />
                    </Dialog.Close>
                </Dialog.Title>
                <div className="p-5 bodyMd">{children}</div>
                {(action || secondaryAction) && <div className='flex flex-row gap-2 p-5 justify-end'>
                    {secondaryAction && <Dialog.Close asChild>
                        <Button content={secondaryAction.content} variant='outline' onClick={secondaryAction.onAction} />
                    </Dialog.Close>}
                    {action && <Button content={action.content} onClick={action.onAction} />}
                </div>}

            </Dialog.Content>
        </Dialog.Portal>
    </Dialog.Root>
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
