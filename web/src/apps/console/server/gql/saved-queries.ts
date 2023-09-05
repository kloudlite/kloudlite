import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGQLServerProps } from '~/root/lib/types/common';
import { accountQueries } from './queries/account-queries';
import { projectQueries } from './queries/project-queries';
import { clusterQueries } from './queries/cluster-queries';
import { providerSecretQueries } from './queries/provider-secret-queries';
import { nodepoolQueries } from './queries/nodepool-queries';
import { workspaceQueries } from './queries/workspace-queries';
import { appQueries } from './queries/app-queries';
import { routerQueries } from './queries/router-queries';
import { configQueries } from './queries/config-queries';
import { secretQueries } from './queries/secret-queries';
import { baseQueries } from './queries/base-queries';
import { environmentQueries } from './queries/environment-queries';

export const GQLServerHandler = ({ headers, cookies }: IGQLServerProps) => {
  const executor = ExecuteQueryWithContext(headers, cookies);
  return {
    ...baseQueries(executor),
    ...accountQueries(executor),
    ...projectQueries(executor),
    ...clusterQueries(executor),
    ...providerSecretQueries(executor),
    ...nodepoolQueries(executor),
    ...workspaceQueries(executor),
    ...environmentQueries(executor),
    ...appQueries(executor),
    ...routerQueries(executor),
    ...configQueries(executor),
    ...secretQueries(executor),
  };
};

export type ConsoleApiType = ReturnType<typeof GQLServerHandler>;
