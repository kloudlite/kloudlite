import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { ensureAccountSet } from '../server/utils/auth-utils';
import { NewCluster } from '../page-components/new-cluster';
import { getPagination } from '../server/utils/common';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  const { data, errors } = await GQLServerHandler(
    ctx.request
  ).listProviderSecrets({
    pagination: getPagination(ctx),
  });

  if (errors) {
    logger.error(errors);
  }

  return {
    providerSecrets: data,
  };
};

const _NewCluster = () => {
  return <NewCluster loader={loader} />;
};

export default _NewCluster;
