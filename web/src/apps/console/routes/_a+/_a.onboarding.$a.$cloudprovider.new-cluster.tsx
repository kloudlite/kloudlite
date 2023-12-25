import { IRemixCtx } from '~/root/lib/types/common';
import { useLoaderData } from '@remix-run/react';
import { defer } from '@remix-run/node';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { NewCluster } from '~/console/page-components/new-cluster';


export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { cloudprovider: cp } = ctx.params;
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).getProviderSecret({
      name: cp,
    });

    if (errors) {
      return { redirect: '/teams', cloudProvider: data };
    }

    return {
      cloudProvider: data,
      redirect: '',
    };
  });
  return defer({ promise });
};

const _NewCluster = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp data={promise}>
      {({ cloudProvider }) => {
        return <NewCluster cloudProvider={cloudProvider} />;
      }}
    </LoadingComp>
  );
};

export default _NewCluster;
