import { Outlet, useOutletContext, useLoaderData } from '@remix-run/react';
import { redirect } from '@remix-run/node';
import { GQLServerHandler } from '../server/gql/saved-queries';

const Cluster = () => {
  const rootContext = useOutletContext();
  const { cluster } = useLoaderData();
  // @ts-ignore
  return <Outlet context={{ ...rootContext, cluster }} />;
};

export const handle = {
  navbar: [
    {
      label: 'Nodepools',
      href: '/nodepools',
      key: 'nodepools',
      value: '/nodepools',
    },

    {
      label: 'Projects',
      href: '/projects',
      key: 'projects',
      value: '/projects',
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
  const { account, cluster } = ctx.params;
  const baseurl = `/${account}/${cluster}`;
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getCluster({
      name: cluster,
    });
    if (errors) {
      throw errors[0];
    }
    return {
      baseurl,
      cluster: data || {},
    };
  } catch (err) {
    return redirect(`/${account}/clusters`);
  }
};

export default Cluster;
