import classNames from 'classnames';
import { motion } from 'framer-motion';
import { FormEventHandler } from 'react';
import { ChildrenProps } from '~/components/types';

export const FadeIn = ({
  children,
  className = '',
  onSubmit,
}: ChildrenProps & {
  className?: string;
  onSubmit?: FormEventHandler<HTMLFormElement>;
}) => {
  return (
    <motion.form
      onSubmit={onSubmit}
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ ease: 'linear', duration: 0.3 }}
      className={classNames(
        'flex flex-col gap-6xl w-full justify-center',
        className
      )}
    >
      {children}
    </motion.form>
  );
};
