import logger from '~/root/lib/client/helpers/log';
import { redirect } from 'react-router-dom';
import { IRemixCtx } from '~/root/lib/types/common';
import { useLoaderData } from '@remix-run/react';
import { defer } from '@remix-run/node';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { ensureAccountSet } from '../server/utils/auth-utils';
import { NewCluster } from '../page-components/new-cluster';
import { LoadingComp, pWrapper } from '../components/loading-component';
import { IProviderSecret } from '../server/gql/queries/provider-secret-queries';

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
      logger.error(errors);
      const cp: IProviderSecret = {} as any;
      return { redirect: '/teams', cloudProvider: cp };
    }

    return {
      cloudProvider: data,
    };
  });
  return defer({ promise });
};

const _NewCluster = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp data={promise}>
      {({ cloudProvider }) => {
        return <NewCluster cloudProvider={cloudProvider as any} />;
      }}
    </LoadingComp>
  );
};

export default _NewCluster;
