import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGQLServerProps } from '~/root/lib/types/common';
import { accessQueries } from './queries/access-queries';
import { accountQueries } from './queries/account-queries';
import { appQueries } from './queries/app-queries';
import { baseQueries } from './queries/base-queries';
import { clusterQueries } from './queries/cluster-queries';
import { configQueries } from './queries/config-queries';
import { crQueries } from './queries/cr-queries';
import { environmentQueries } from './queries/environment-queries';
import { managedResourceQueries } from './queries/managed-resource-queries';
import { managedServiceQueries } from './queries/managed-service-queries';
import { nodepoolQueries } from './queries/nodepool-queries';
import { projectQueries } from './queries/project-queries';
import { providerSecretQueries } from './queries/provider-secret-queries';
import { repoQueries } from './queries/repo-queries';
import { routerQueries } from './queries/router-queries';
import { secretQueries } from './queries/secret-queries';
import { tagsQueries } from './queries/tags-queries';
import { vpnQueries } from './queries/vpn-queries';
import { workspaceQueries } from './queries/workspace-queries';

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
    ...vpnQueries(executor),
    ...accessQueries(executor),
    ...managedServiceQueries(executor),
    ...managedResourceQueries(executor),
    ...crQueries(executor),
    ...repoQueries(executor),
    ...tagsQueries(executor),
  };
};

export type ConsoleApiType = ReturnType<typeof GQLServerHandler>;
