import NewProject from '~/console/page-components/new-project';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';


const _NewProject = () => {
  return <NewProject />;
};

export const loader = async (ctx: IRemixCtx) => {
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

export type INewProjectOnBoardingLoader = ReturnType<typeof loader>;

export default _NewProject;
