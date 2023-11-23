import { ReactNode } from 'react';
import { ProdLogo } from '~/components/branding/prod-logo';
import { WorkspacesLogo } from '~/components/branding/workspace-logo';

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
