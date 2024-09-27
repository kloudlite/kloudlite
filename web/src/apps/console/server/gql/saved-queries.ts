import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGQLServerProps } from '~/root/lib/types/common';
import { accessQueries } from './queries/access-queries';
import { accountQueries } from './queries/account-queries';
import { appQueries } from './queries/app-queries';
import { baseQueries } from './queries/base-queries';
import { buildQueries } from './queries/build-queries';
import { buildRunQueries } from './queries/build-run-queries';
import { byokClusterQueries } from './queries/byok-cluster-queries';
import { clusterManagedServicesQueries } from './queries/cluster-managed-services-queries';
import { clusterQueries } from './queries/cluster-queries';
import { commsNotificationQueries } from './queries/comms-queries';
import { configQueries } from './queries/config-queries';
import { crQueries } from './queries/cr-queries';
import { domainQueries } from './queries/domain-queries';
import { environmentQueries } from './queries/environment-queries';
import { externalAppQueries } from './queries/external-app-queries';
import { gitQueries } from './queries/git-queries';
import { globalVpnQueries } from './queries/global-vpn-queries';
import { helmChartQueries } from './queries/helm-chart-queries';
import { imagePullSecretsQueries } from './queries/image-pull-secrets-queries';
import { importedManagedResourceQueries } from './queries/imported-managed-resource-queries';
import { managedResourceQueries } from './queries/managed-resources-queries';
import { managedTemplateQueries } from './queries/managed-templates-queries';
import { namespaceQueries } from './queries/namespace-queries';
import { nodepoolQueries } from './queries/nodepool-queries';
import { providerSecretQueries } from './queries/provider-secret-queries';
import { pvQueries } from './queries/pv-queries';
import { pvcQueries } from './queries/pvc-queries';
import { registryImagesQueries } from './queries/registry-image-queries';
import { repoQueries } from './queries/repo-queries';
import { routerQueries } from './queries/router-queries';
import { secretQueries } from './queries/secret-queries';
import { tagsQueries } from './queries/tags-queries';

export const GQLServerHandler = ({ headers, cookies }: IGQLServerProps) => {
  const executor = ExecuteQueryWithContext(headers, cookies);
  return {
    ...baseQueries(executor),
    ...accountQueries(executor),
    ...clusterQueries(executor),
    ...providerSecretQueries(executor),
    ...nodepoolQueries(executor),
    ...environmentQueries(executor),
    ...appQueries(executor),
    ...externalAppQueries(executor),
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
    ...managedTemplateQueries(executor),
    ...managedResourceQueries(executor),
    ...helmChartQueries(executor),
    ...namespaceQueries(executor),
    ...imagePullSecretsQueries(executor),
    ...globalVpnQueries(executor),
    ...commsNotificationQueries(executor),
    ...importedManagedResourceQueries(executor),
    ...registryImagesQueries(executor),
  };
};

export type ConsoleApiType = ReturnType<typeof GQLServerHandler>;
