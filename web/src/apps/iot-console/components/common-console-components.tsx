import { ReactNode, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { cn } from '~/components/utils';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { toast } from '~/components/molecule/toast';
import { Copy, Check } from '~/iotconsole/components/icons';
import { ListBody } from './console-list-components';

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
  className?: string;
}

export const Box = ({ children, title, className }: IBox) => {
  return (
    <div
      className={cn(
        'rounded border border-border-default bg-surface-basic-default shadow-button p-3xl flex flex-col gap-3xl ',
        className
      )}
    >
      <div className="text-text-strong headingLg">{title}</div>
      <div className="flex flex-col gap-3xl flex-1">{children}</div>
    </div>
  );
};

export const CopyContentToClipboard = ({
  content,
  toastMessage,
  label,
}: {
  content: string;
  toastMessage: string;
  label?: ReactNode;
}) => {
  const iconSize = 16;
  const { copy } = useClipboard({});
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    copy(content);
    setCopied(true);
    toast.success(toastMessage);

    setTimeout(() => {
      setCopied(false);
    }, 3000);
  };

  return (
    <ListBody
      data={
        <div
          className="cursor-pointer flex flex-row items-center gap-lg truncate hover:text-text-default"
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            if (!copied) {
              handleCopy();
            }
          }}
        >
          <span className="truncate">{label || content}</span>
          {copied ? (
            <span>
              <Check size={iconSize} />
            </span>
          ) : (
            <span>
              <Copy size={iconSize} />
            </span>
          )}
        </div>
      }
    />
  );
};
