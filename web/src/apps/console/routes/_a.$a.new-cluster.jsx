import logger from '~/root/lib/client/helpers/log';
import { getPagination } from '../server/r-urils/common';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { ensureAccountSet } from '../server/utils/auth-utils';
import { NewCluster } from '../page-components/new-cluster';

const _NewCluster = () => {
  return <NewCluster />;
};

export const loader = async (ctx) => {
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

export default _NewCluster;
