import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGQLServerProps } from '~/root/lib/types/common';
import { accessQueries } from './queries/access-queries';
import { accountQueries } from './queries/account-queries';
import { appQueries } from './queries/app-queries';
import { baseQueries } from './queries/base-queries';
import { buildQueries } from './queries/build-queries';
import { buildRunQueries } from './queries/build-run-queries';
import { clusterQueries } from './queries/cluster-queries';
import { configQueries } from './queries/config-queries';
import { crQueries } from './queries/cr-queries';
import { domainQueries } from './queries/domain-queries';
import { environmentQueries } from './queries/environment-queries';
import { gitQueries } from './queries/git-queries';
import { nodepoolQueries } from './queries/nodepool-queries';
import { byokClusterQueries } from './queries/byok-cluster-queries';
// import { projectQueries } from './queries/project-queries';
import { providerSecretQueries } from './queries/provider-secret-queries';
import { repoQueries } from './queries/repo-queries';
import { routerQueries } from './queries/router-queries';
import { secretQueries } from './queries/secret-queries';
import { tagsQueries } from './queries/tags-queries';
import { pvcQueries } from './queries/pvc-queries';
import { pvQueries } from './queries/pv-queries';
import { clusterManagedServicesQueries } from './queries/cluster-managed-services-queries';
// import { projectManagedServicesQueries } from './queries/project-managed-services-queries';
import { managedResourceQueries } from './queries/managed-resources-queries';
import { managedTemplateQueries } from './queries/managed-templates-queries';
import { helmChartQueries } from './queries/helm-chart-queries';
import { namespaceQueries } from './queries/namespace-queries';
import { consoleVpnQueries } from './queries/console-vpn-queries';
import { imagePullSecretsQueries } from './queries/image-pull-secrets-queries';

export const GQLServerHandler = ({ headers, cookies }: IGQLServerProps) => {
  const executor = ExecuteQueryWithContext(headers, cookies);
  return {
    ...baseQueries(executor),
    ...accountQueries(executor),
    // ...projectQueries(executor),
    ...clusterQueries(executor),
    ...providerSecretQueries(executor),
    ...nodepoolQueries(executor),
    ...environmentQueries(executor),
    ...appQueries(executor),
    ...routerQueries(executor),
    ...configQueries(executor),
    ...secretQueries(executor),
    ...accessQueries(executor),
    ...crQueries(executor),
    ...repoQueries(executor),
    ...tagsQueries(executor),
    ...gitQueries(executor),
    ...domainQueries(executor),
    ...buildQueries(executor),
    ...pvcQueries(executor),
    ...pvQueries(executor),
    ...buildRunQueries(executor),
    ...clusterManagedServicesQueries(executor),
    ...byokClusterQueries(executor),
    // ...projectManagedServicesQueries(executor),
    ...managedTemplateQueries(executor),
    ...managedResourceQueries(executor),
    ...helmChartQueries(executor),
    ...namespaceQueries(executor),
    ...consoleVpnQueries(executor),
    ...imagePullSecretsQueries(executor),
  };
};

export type ConsoleApiType = ReturnType<typeof GQLServerHandler>;
