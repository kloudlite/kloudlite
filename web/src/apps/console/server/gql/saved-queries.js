import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';
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

export const GQLServerHandler = ({ headers, cookies }) => {
  const executor = ExecuteQueryWithContext(headers, cookies);
  return {
    ...accountQueries(executor),
    ...projectQueries(executor),
    ...clusterQueries(executor),
    ...providerSecretQueries(executor),
    ...nodepoolQueries(executor),
    ...workspaceQueries(executor),
    ...appQueries(executor),
    ...routerQueries(executor),
    ...configQueries(executor),
    ...secretQueries(executor),

    infraCheckNameAvailability: executor(
      gql`
        query Infra_checkNameAvailability($resType: ResType!, $name: String!) {
          infra_checkNameAvailability(resType: $resType, name: $name) {
            suggestedNames
            result
          }
        }
      `,
      {
        dataPath: 'infra_checkNameAvailability',
      }
    ),

    coreCheckNameAvailability: executor(
      gql`
        query Core_checkNameAvailability(
          $resType: ConsoleResType!
          $name: String!
          $namespace: String
        ) {
          core_checkNameAvailability(
            resType: $resType
            name: $name
            namespace: $namespace
          ) {
            result
          }
        }
      `,
      {
        dataPath: 'core_checkNameAvailability',
      }
    ),

    logout: executor(gql`
      mutation Auth {
        auth_logout
      }
    `),

    whoAmI: executor(
      gql`
        query Auth_me {
          auth_me {
            id
            email
          }
        }
      `,
      {
        dataPath: 'auth_me',
        transformer: (me) => ({ me }),
      }
    ),
  };
};
