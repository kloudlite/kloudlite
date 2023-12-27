import { IRemixCtx } from '~/root/lib/types/common';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { getPagination } from '~/console/server/utils/common';
import { NewCluster } from '~/console/page-components/new-cluster';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listProviderSecrets({
      pagination: getPagination(ctx),
    });

    if (errors) {
      throw errors[0];
    }

    return {
      providerSecrets: data,
    };
  });
  return defer({ promise });
};

const _NewCluster = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp data={promise}>
      {({ providerSecrets }) => {
        return <NewCluster providerSecrets={providerSecrets} />;
      }}
    </LoadingComp>
  );
};

export default _NewCluster;
