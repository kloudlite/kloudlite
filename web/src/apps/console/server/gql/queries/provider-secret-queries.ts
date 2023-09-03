import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  ConsoleListProviderSecretsQuery,
  ConsoleListProviderSecretsQueryVariables,
} from '~/root/src/generated/gql/server';

export const providerSecretQueries = (executor: IExecutor) => ({
  listProviderSecrets: executor(
    gql`
      query Edgess(
        $pagination: CursorPaginationIn
        $search: SearchProviderSecret
      ) {
        infra_listProviderSecrets(pagination: $pagination, search: $search) {
          edges {
            node {
              enabled
              stringData
              metadata {
                annotations
                name
              }

              cloudProviderName
              status {
                resources {
                  namespace
                  name
                  kind
                  apiVersion
                }
                message {
                  RawMessage
                }
                lastReconcileTime
                isReady
                checks
              }
              creationTime
              updateTime
            }
          }

          totalCount
          pageInfo {
            startCursor
            hasPreviousPage
            hasNextPage
            endCursor
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleListProviderSecretsQuery) => {
        return data.infra_listProviderSecrets;
      },
      vars: (_: ConsoleListProviderSecretsQueryVariables) => {},
    }
  ),
  createProviderSecret: executor(
    gql`
      mutation createProviderSecret($secret: CloudProviderSecretIn!) {
        infra_createProviderSecret(secret: $secret) {
          metadata {
            name
          }
        }
      }
    `,
    {
      transformer(data) {},
      vars(variables) {},
    }
  ),
  updateProviderSecret: executor(
    gql`
      mutation Infra_updateProviderSecret($secret: CloudProviderSecretIn!) {
        infra_updateProviderSecret(secret: $secret) {
          id
        }
      }
    `,
    {
      transformer: (data) => data,
      vars(_) {},
    }
  ),
  deleteProviderSecret: executor(
    gql`
      mutation Infra_deleteProviderSecret($secretName: String!) {
        infra_deleteProviderSecret(secretName: $secretName)
      }
    `,
    {
      transformer(data) {},
      vars(variables) {},
    }
  ),
  getProviderSecret: executor(
    gql`
      query Metadata($name: String!) {
        infra_getProviderSecret(name: $name) {
          metadata {
            name
            annotations
          }
          cloudProviderName
        }
      }
    `,

    {
      transformer(data) {},
      vars(variables) {},
    }
  ),
});
