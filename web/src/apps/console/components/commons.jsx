import { ProdLogo } from '~/components/branding/prod-logo';
import { WorkspacesLogo } from '~/components/branding/workspace-logo';

export const BlackProdLogo = ({ size = 16 }) => {
  return <ProdLogo color="currentColor" size={size} />;
};

export const BlackWorkspaceLogo = ({ size = 16 }) => {
  return <WorkspacesLogo color="currentColor" size={size} />;
};
