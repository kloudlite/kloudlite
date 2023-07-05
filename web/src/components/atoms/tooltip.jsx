import React, { useState } from 'react';
import * as T from '@radix-ui/react-tooltip';
import { AnimatePresence, motion } from 'framer-motion';


export const TooltipProvider = ({ delayDuration = 0, children }) => (
    <T.Provider delayDuration={delayDuration}>
        {children}
    </T.Provider>
)

export const Tooltip = ({ children, content }) => {
    const [open, setOpen] = useState(false)
    return (
        <T.Root open={open} onOpenChange={setOpen}>
            <T.Trigger asChild>
                {children}
            </T.Trigger>

            <AnimatePresence>
                {open &&
                    <T.Portal forceMount>
                        <T.Content
                            asChild
                            sideOffset={5}
                        >
                            <motion.div
                                initial={{ y: -2, opacity: 0 }}
                                animate={{ y: 0, opacity: 1 }}
                                exit={{ y: -2, opacity: 0 }}
                                transition={{ duration: 0.3, ease: 'anticipate' }}
                                className="bodySm-default text-text-default px-2 py-1 shadow-popover bg-surface-default rounded"
                            >
                                {content}
                            </motion.div>
                        </T.Content>
                    </T.Portal>
                }

            </AnimatePresence>

        </T.Root>
    );
};

export default Tooltip;