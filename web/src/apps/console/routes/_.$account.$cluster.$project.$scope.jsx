import { Outlet, useOutletContext, Link } from '@remix-run/react';
import { ensureAccountSet, ensureClusterSet } from '../server/utils/auth-utils';
import Breadcrum from '../components/breadcrum';

const Project = () => {
  const rootContext = useOutletContext();
  return <Outlet context={{ ...rootContext }} />;
};

export default Project;

export const handle = ({ account, cluster, project, scope }) => {
  return {
    breadcrum: () => (
      <Breadcrum.Button
        content={project}
        LinkComponent={Link}
        href={`/${account}/${cluster}/${project}/${
          scope === 'workspace' ? 'workspaces' : 'environments'
        }`}
      />
    ),
  };
};

export const loader = async (ctx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { account, project, cluster, scope } = ctx.params;
  const baseurl = `/${account}/${cluster}/${project}`;

  return {
    baseurl,
    account,
    cluster,
    project,
    scope,
  };
};
