import logger from '~/root/lib/client/helpers/log';
import { ensureAccountSet } from '../server/utils/auth-utils';
import { GQLServerHandler } from '../server/gql/saved-queries';
import NewProject from '../page-components/new-project';

const _NewProject = () => {
  return <NewProject />;
};

export const loader = async (ctx) => {
  const { cluster } = ctx.params;
  ensureAccountSet(ctx);
  const { data, errors } = await GQLServerHandler(ctx.request).getCluster({
    name: cluster,
  });

  if (errors) {
    logger.error(errors);
  }

  return {
    cluster: data,
  };
};

export default _NewProject;
