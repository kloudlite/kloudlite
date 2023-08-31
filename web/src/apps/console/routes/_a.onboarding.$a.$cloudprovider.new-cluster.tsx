import logger from '~/root/lib/client/helpers/log';
import { redirect } from 'react-router-dom';
import { IRCtx } from '~/root/lib/types/common';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { ensureAccountSet } from '../server/utils/auth-utils';
import { NewCluster } from '../page-components/new-cluster';

const _NewCluster = () => {
  return <NewCluster />;
};

export const loader = async (ctx: IRCtx) => {
  ensureAccountSet(ctx);
  const { cloudprovider: cp } = ctx.params;
  const { data, errors } = await GQLServerHandler(
    ctx.request
  ).getProviderSecret({
    name: cp,
  });

  if (errors) {
    logger.error(errors);
    return redirect('/teams');
  }

  return {
    cloudProvider: data,
  };
};

export default _NewCluster;
