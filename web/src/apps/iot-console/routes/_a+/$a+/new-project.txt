import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import NewProject from '~/iotconsole/page-components/new-project';
import { ensureAccountSet } from '~/iotconsole/server/utils/auth-utils';
import { getPagination, getSearch } from '~/iotconsole/server/utils/common';

const _NewProject = () => {
  return <NewProject />;
};

export default _NewProject;
