import { ReactNode, useState } from 'react';
import { Button } from '@kloudlite/design-system/atoms/button';
import TooltipV2 from '@kloudlite/design-system/atoms/tooltipV2';
import { toast } from '@kloudlite/design-system/molecule/toast';
import { cn } from '@kloudlite/design-system/utils';
import { Check, Copy } from '~/console/components/icons';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { Truncate } from '~/root/lib/utils/common';

interface IDeleteContainer {
  title: ReactNode;
  children: ReactNode;
  action: () => void;
  content?: string;
  disabled?: boolean;
}
export const DeleteContainer = ({
  title,
  children,
  action,
  content,
  disabled = false,
}: IDeleteContainer) => {
  return (
    <div className="flex flex-col gap-3xl p-3xl rounded border border-border-critical bg-surface-basic-default shadow-button">
      <div className="text-text-strong headingLg">{title}</div>
      <div className="bodyMd text-text-default">{children}</div>
      <Button
        onClick={action}
        content={content || 'Delete'}
        variant="critical"
        disabled={disabled}
      />
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
  truncateLength = 20,
}: {
  content: string;
  toastMessage: string;
  truncateLength?: number;
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
      className="flex flex-row items-center flex-1 cursor-pointer group"
      onClick={(e) => {
        e.preventDefault();
        e.stopPropagation();
        if (!copied) {
          handleCopy();
        }
      }}
    >
      <span className="bodyMd-medium text-text-strong">
        {content.length >= truncateLength ? (
          <TooltipV2 content={content}>
            <Truncate length={truncateLength}>{content}</Truncate>
          </TooltipV2>
        ) : (
          content
        )}
      </span>
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
