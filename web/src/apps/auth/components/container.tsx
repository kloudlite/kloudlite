import { Link } from '@remix-run/react';
import { ReactNode } from 'react';
import { Button } from '~/components/atoms/button';
import { cn } from '~/components/utils';

interface ContainerProps {
  children: ReactNode;
  footer?: {
    message: string;
    buttonText: string;
    to: string;
  };
}

const Container = ({ children, footer }: ContainerProps) => {
  return (
    <div className={cn('flex flex-col items-center justify-start h-full')}>
      <div
        className={cn(
          'flex flex-1 flex-col items-center self-stretch justify-center px-3xl py-10xl'
        )}
      >
        {children}
      </div>
      {footer && (
        <div className="py-5xl px-3xl flex flex-row items-center justify-center self-stretch border-t border-border-default sticky bottom-0 bg-surface-basic-default">
          <div className="bodyMd text-text-default">{footer?.message}</div>
          <Button
            content={footer?.buttonText}
            variant="primary-plain"
            size="md"
            to={footer?.to}
            LinkComponent={Link}
          />
        </div>
      )}
    </div>
  );
};

export default Container;
