import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { ensureAccountSet } from '../server/utils/auth-utils';
import { NewCluster } from '../page-components/new-cluster';
import { getPagination } from '../server/utils/common';
import { LoadingComp, pWrapper } from '../components/loading-component';

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
        if (!providerSecrets) return null;
        return <NewCluster providerSecrets={providerSecrets as any} />;
      }}
    </LoadingComp>
  );
};

export default _NewCluster;
