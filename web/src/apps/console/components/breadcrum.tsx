import React, { ReactNode } from 'react';
import { ButtonProps, Button as NativeButton } from '~/components/atoms/button';

interface RootProps {
  children: ReactNode;
}

const Root = ({ children }: RootProps) => {
  return <div className="flex flex-row gap-md items-center">{children}</div>;
};

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (props, ref) => {
    return (
      <div className="flex flex-row gap-md items-center">
        <div className="text-text-disabled bodySm">/</div>
        <NativeButton size="md" variant="plain" ref={ref} {...props} />
      </div>
    );
  }
);

const Breadcrum = {
  Root,
  Button,
};
export default Breadcrum;
