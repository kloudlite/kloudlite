import { redirect } from '@remix-run/node';
import { Link, Outlet, useOutletContext, useParams } from '@remix-run/react';
import { ProdLogo } from '~/components/branding/prod-logo';
import { WorkspacesLogo } from '~/components/branding/workspace-logo';
import { IRemixCtx } from '~/root/lib/types/common';
import Breadcrum from '../components/breadcrum';
import LogoWrapper from '../components/logo-wrapper';
import { SCOPE } from '../page-components/new-scope';
import { ensureAccountSet, ensureClusterSet } from '../server/utils/auth-utils';
import { IProjectContext } from './_.$account.infra.$cluster.$project';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { account, cluster, project, scope } = ctx.params;
  switch (scope) {
    case SCOPE.ENVIRONMENT:
    case SCOPE.WORKSPACE:
      return {};
    default:
      return redirect(`/${account}/${cluster}/${project}/environments`);
  }
};

const Project = () => {
  const rootContext = useOutletContext<IProjectContext>();
  return <Outlet context={{ ...rootContext }} />;
};

export default Project;

const ScopeBreadcrumButton = () => {
  const { account, cluster, project, scope } = useParams();
  return (
    <Breadcrum.Button
      content={project}
      LinkComponent={Link}
      to={`/${account}/${cluster}/${project}/${
        scope === 'workspace' ? 'workspaces' : 'environments'
      }`}
    />
  );
};

const BrandLogo = () => {
  const { scope } = useParams();
  return scope === 'workspace' ? <WorkspacesLogo /> : <ProdLogo />;
};

const Logo = () => {
  const { account, cluster, project, scope, workspace } = useParams();
  return (
    <LogoWrapper
      to={`/${account}/${cluster}/${project}/${scope}/${workspace}/apps`}
    >
      <BrandLogo />
    </LogoWrapper>
  );
};

export const handle = () => {
  return {
    breadcrum: () => <ScopeBreadcrumButton />,
    logo: <Logo />,
  };
};
