import { ReactNode, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { cn } from '~/components/utils';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { toast } from '~/components/molecule/toast';
import { Copy, Check } from '~/console/components/icons';
import { ListItem } from './console-list-components';

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
  toolTip = false,
}: {
  content: string;
  toastMessage: string;
  label?: ReactNode;
  toolTip?: boolean;
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
    <div
      className="flex flex-row items-center truncate flex-1 cursor-pointer group"
      onClick={(e) => {
        e.preventDefault();
        e.stopPropagation();
        if (!copied) {
          handleCopy();
        }
      }}
    >
      <ListItem
        className="flex-1"
        noTooltip={!toolTip}
        data={
          <span className="cursor-pointer items-center gap-lg hover:text-text-default group-hover:text-text-default group-[.is-data]:truncate">
            {label || content}
          </span>
        }
      />
      <div className="shrink-0 ml-md">
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
    </div>
  );
};
