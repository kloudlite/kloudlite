import { Outlet, useOutletContext, useLoaderData } from '@remix-run/react';
import { redirect } from '@remix-run/node';
import { GQLServerHandler } from '../server/gql/saved-queries';

const Project = () => {
  const rootContext = useOutletContext();
  const { project } = useLoaderData();
  // @ts-ignore
  return <Outlet context={{ ...rootContext, project }} />;
};

export default Project;

export const handle = {
  navbar: [
    {
      label: 'Workspaces',
      href: '/workspaces',
      key: 'workspaces',
      value: '/workspaces',
    },
    {
      label: 'Environments',
      href: '/environments',
      key: 'environments',
      value: '/environments',
    },
    {
      label: 'Settings',
      href: '/settings',
      key: 'settings',
      value: '/settings',
    },
  ],
};
export const loader = async (ctx) => {
  const { account, project, cluster } = ctx.params;
  const baseurl = `/${account}/${cluster}/${project}`;
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getProject({
      name: project,
    });
    if (errors) {
      throw errors[0];
    }
    return {
      baseurl,
      project: data || {},
    };
  } catch (err) {
    return redirect(`/${account}/${cluster}/projects`);
  }
};
