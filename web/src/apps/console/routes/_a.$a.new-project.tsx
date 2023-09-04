import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import { ensureAccountSet } from '../server/utils/auth-utils';
import { GQLServerHandler } from '../server/gql/saved-queries';
import NewProject from '../page-components/new-project';
import { getPagination, getSearch } from '../server/utils/common';

const _NewProject = () => {
  return <NewProject />;
};

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  const { data, errors } = await GQLServerHandler(ctx.request).listClusters({
    pagination: getPagination(ctx),
    search: getSearch(ctx),
  });

  if (errors) {
    logger.error(errors);
  }

  return {
    clustersData: data,
  };
};

export default _NewProject;
