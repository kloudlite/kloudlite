import { redirect } from '@remix-run/node';
import {
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import withContext from '~/root/lib/app-setup/with-contxt';
import { IExtRemixCtx } from '~/root/lib/types/common';
import { CommonTabs } from '../components/common-navbar-tabs';
import { type ICluster } from '../server/gql/queries/cluster-queries';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { ensureAccountSet, ensureClusterSet } from '../server/utils/auth-utils';
import { IAccountContext } from './_.$account';

export interface IClusterContext extends IAccountContext {
  cluster: ICluster;
}

const Cluster = () => {
  const rootContext = useOutletContext<IAccountContext>();
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
          value: '/nodepools',
        },
        {
          label: 'Settings',
          to: '/settings',
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

export const loader = async (ctx: IExtRemixCtx) => {
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
      cluster: data,
    });
  } catch (err) {
    return redirect(`/${account}/clusters`);
  }
};

export default Cluster;
