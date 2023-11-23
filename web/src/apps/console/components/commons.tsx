import { CopySimple, Check } from '@jengaicons/react';
import { ReactNode, useState } from 'react';
import Chips from '~/components/atoms/chips';
import { ProdLogo } from '~/components/branding/prod-logo';
import { WorkspacesLogo } from '~/components/branding/workspace-logo';
import { toast } from '~/components/molecule/toast';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';

export const BlackProdLogo = ({ size = 16 }) => {
  return <ProdLogo color="currentColor" size={size} />;
};

export const BlackWorkspaceLogo = ({ size = 16 }) => {
  return <WorkspacesLogo color="currentColor" size={size} />;
};

export const DetailItem = ({
  title,
  value,
}: {
  title: ReactNode;
  value: ReactNode;
}) => {
  return (
    <div className="flex flex-col gap-lg flex-1 min-w-[45%]">
      <div className="bodyMd-medium text-text-default">{title}</div>
      <div className="bodyMd text-text-strong">{value}</div>
    </div>
  );
};

export const CopyButton = ({
  title,
  value,
}: {
  title: ReactNode;
  value: string;
}) => {
  const [copyIcon, setCopyIcon] = useState(<CopySimple />);
  const { copy } = useClipboard({
    onSuccess: () => {
      setTimeout(() => {
        setCopyIcon(<CopySimple />);
      }, 1000);
      toast.success('Copied to clipboard');
    },
  });

  return (
    <Chips.Chip
      type="CLICKABLE"
      item={title}
      label={title}
      prefix={copyIcon}
      onClick={() => {
        copy(value);
        setCopyIcon(<Check />);
      }}
    />
  );
};
