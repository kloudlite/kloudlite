import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import { getPagination } from '../server/r-urils/common';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { ensureAccountSet } from '../server/utils/auth-utils';
import { NewCluster } from '../page-components/new-cluster';

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
