import React, { ReactNode } from 'react';
import { IButton, Button as NativeButton } from '~/components/atoms/button';

interface IBreadcrum {
  children: ReactNode;
}

const Root = ({ children }: IBreadcrum) => {
  return <div className="flex flex-row gap-md items-center">{children}</div>;
};

const Button = React.forwardRef<HTMLButtonElement, IButton>((props, ref) => {
  return (
    <div className="flex flex-row gap-md items-center bodyMd-medium">
      <NativeButton size="md" variant="plain" ref={ref} {...props} />
    </div>
  );
});

const Breadcrum = {
  Root,
  Button,
};
export default Breadcrum;
