import * as Dialog from '@radix-ui/react-dialog';
import { AnimatePresence, motion } from 'framer-motion';
import { ReactNode } from 'react';
import { cn } from '~/components/utils';

const OverlaySideDialog = ({
  show = false,
  onOpenChange = () => {},
  children,
  backdrop = true,
  className = '',
}: {
  show?: boolean;
  onOpenChange?: (val: any) => void;
  backdrop?: boolean;
  className?: string;
  children?: ReactNode;
}) => {
  return (
    <Dialog.Root
      open={show}
      onOpenChange={(e) => {
        if (e) {
          onOpenChange(show);
        } else {
          onOpenChange(false);
        }
      }}
    >
      <AnimatePresence>
        {show && (
          <Dialog.Portal forceMount>
            <Dialog.Overlay asChild forceMount>
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.3, ease: 'linear' }}
                className={cn('fixed inset-0 z-[9999999]', {
                  'bg-text-default/60': backdrop,
                })}
              />
            </Dialog.Overlay>
            <Dialog.Content asChild forceMount>
              <motion.div
                initial={{ x: '75%', opacity: 1 }}
                animate={{ x: '0%', opacity: 1 }}
                exit={{ x: '100%', opacity: 1 }}
                transition={{ duration: 0.3, ease: 'linear' }}
                className={cn(
                  'flex flex-col',
                  'z-[99999999] outline-none transform overflow-hidden md:rounded bg-surface-basic-default shadow-modal',
                  'fixed right-0 top-0 h-screen w-[50vw] max-w-screen-2xl',
                  'border border-border-default',
                  className
                )}
              >
                {children}
              </motion.div>
            </Dialog.Content>
          </Dialog.Portal>
        )}
      </AnimatePresence>
    </Dialog.Root>
  );
};

export default OverlaySideDialog;
