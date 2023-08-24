import { Outlet, useOutletContext, Link, useParams } from '@remix-run/react';
import { redirect } from '@remix-run/node';
import { ensureAccountSet, ensureClusterSet } from '../server/utils/auth-utils';
import Breadcrum from '../components/breadcrum';
import { SCOPE } from '../page-components/new-scope';

const Project = () => {
  const rootContext = useOutletContext();
  return <Outlet context={{ ...rootContext }} />;
};

export default Project;

const ScopeBreadcrumButton = () => {
  const { account, cluster, project, scope } = useParams();
  return (
    <Breadcrum.Button
      content={project}
      LinkComponent={Link}
      href={`/${account}/${cluster}/${project}/${
        scope === 'workspace' ? 'workspaces' : 'environments'
      }`}
    />
  );
};

export const handle = () => {
  return {
    breadcrum: () => <ScopeBreadcrumButton />,
  };
};

export const loader = async (ctx) => {
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
