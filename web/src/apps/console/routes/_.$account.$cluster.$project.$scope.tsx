import { Outlet, useOutletContext, Link, useParams } from '@remix-run/react';
import { redirect } from '@remix-run/node';
import { IRemixCtx } from '~/root/lib/types/common';
import { ProdLogo } from '~/components/branding/prod-logo';
import { WorkspacesLogo } from '~/components/branding/workspace-logo';
import { ensureAccountSet, ensureClusterSet } from '../server/utils/auth-utils';
import Breadcrum from '../components/breadcrum';
import { SCOPE } from '../page-components/new-scope';
import { IProjectContext } from './_.$account.$cluster.$project';

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

const Logo = () => {
  const { scope } = useParams();
  return scope === 'workspace' ? <WorkspacesLogo /> : <ProdLogo />;
};

const LogoLink = () => {
  const { account, cluster, project, scope, workspace } = useParams();
  return (
    <Link
      to={`/${account}/${cluster}/${project}/${scope}/${workspace}/apps`}
      prefetch="intent"
    >
      <Logo />
    </Link>
  );
};

export const handle = () => {
  return {
    breadcrum: () => <ScopeBreadcrumButton />,
    logo: <LogoLink />,
  };
};
