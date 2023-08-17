import logger from '~/root/lib/client/helpers/log';
import { ensureAccountSet } from '../server/utils/auth-utils';
import { getPagination, getSearch } from '../server/r-urils/common';
import { GQLServerHandler } from '../server/gql/saved-queries';
import NewProject from '../page-components/new-project';

const _NewProject = () => {
  return <NewProject />;
};

export const loader = async (ctx) => {
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
