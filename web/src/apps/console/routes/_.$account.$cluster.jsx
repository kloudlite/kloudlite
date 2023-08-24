import {
  Outlet,
  useOutletContext,
  useLoaderData,
  useParams,
} from '@remix-run/react';
import { redirect } from '@remix-run/node';
import withContext from '~/root/lib/app-setup/with-contxt';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { CommonTabs } from '../components/common-navbar-tabs';
import { ensureAccountSet, ensureClusterSet } from '../server/utils/auth-utils';

const Cluster = () => {
  const rootContext = useOutletContext();
  const { cluster } = useLoaderData();
  return <Outlet context={{ ...rootContext, cluster }} />;
};

const ClusterTabs = () => {
  const { account, cluster } = useParams();
  return (
    <CommonTabs
      tabs={[
        {
          label: 'Nodepools',
          to: '/nodepools',
          key: 'nodepools',
          value: '/nodepools',
        },

        {
          label: 'Projects',
          to: '/projects',
          key: 'projects',
          value: '/projects',
        },
        {
          label: 'Settings',
          to: '/settings',
          key: 'settings',
          value: '/settings',
        },
      ]}
      baseurl={`/${account}/${cluster}`}
      backButton={{
        to: `${account}/clusters`,
        label: 'Clusters',
      }}
    />
  );
};

export const handle = () => {
  return {
    navbar: <ClusterTabs />,
  };
};

export const loader = async (ctx) => {
  const { account, cluster } = ctx.params;
  ensureAccountSet(ctx);
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getCluster({
      name: cluster,
    });
    if (errors) {
      throw errors[0];
    }
    ensureClusterSet(ctx);
    return withContext(ctx, {
      cluster: data || {},
    });
  } catch (err) {
    return redirect(`/${account}/clusters`);
  }
};

export default Cluster;
