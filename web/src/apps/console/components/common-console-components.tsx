import { ReactNode } from 'react';
import { Button } from '~/components/atoms/button';

interface IDeleteContainer {
  title: ReactNode;
  children: ReactNode;
  action: () => void;
}
export const DeleteContainer = ({
  title,
  children,
  action,
}: IDeleteContainer) => {
  return (
    <div className="flex flex-col gap-3xl p-3xl rounded border border-border-critical bg-surface-basic-default shadow-button">
      <div className="text-text-strong headingLg">{title}</div>
      <div className="bodyMd text-text-default">{children}</div>
      <Button onClick={action} content="Delete" variant="critical" />
    </div>
  );
};

interface IBoxPrimitive {
  children: ReactNode;
}

export const BoxPrimitive = ({ children }: IBoxPrimitive) => {
  return (
    <div className="rounded border border-border-default bg-surface-basic-default shadow-button p-3xl flex flex-col gap-3xl">
      {children}
    </div>
  );
};

interface IBox extends IBoxPrimitive {
  title: ReactNode;
}

export const Box = ({ children, title }: IBox) => {
  return (
    <div className="rounded border border-border-default bg-surface-basic-default shadow-button p-3xl flex flex-col gap-3xl ">
      <div className="text-text-strong headingLg">{title}</div>
      <div className="flex flex-col gap-3xl">{children}</div>
    </div>
  );
};
