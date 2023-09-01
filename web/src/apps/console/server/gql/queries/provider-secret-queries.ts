import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGqlReturn } from '~/root/lib/types/common';

export interface IGQLMethodsProviderSecret {
  listProviderSecrets: (variables?: any) => IGqlReturn<any>;
  createProviderSecret: (variables?: any) => IGqlReturn<any>;
  updateProviderSecret: (variables?: any) => IGqlReturn<any>;
  deleteProviderSecret: (variables?: any) => IGqlReturn<any>;
  getProviderSecret: (variables?: any) => IGqlReturn<any>;
}

export const providerSecretQueries = (
  executor: IExecutor
): IGQLMethodsProviderSecret => ({
  listProviderSecrets: executor(
    gql`
      query Edges(
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
      dataPath: 'infra_listProviderSecrets',
    }
  ),
  createProviderSecret: executor(
    gql`
      mutation Mutation($secret: CloudProviderSecretIn!) {
        infra_createProviderSecret(secret: $secret) {
          metadata {
            name
          }
        }
      }
    `,
    {
      dataPath: 'infra_createProviderSecret',
    }
  ),
  updateProviderSecret: executor(gql`
    mutation Infra_updateProviderSecret($secret: CloudProviderSecretIn!) {
      infra_updateProviderSecret(secret: $secret) {
        id
      }
    }
  `),
  deleteProviderSecret: executor(gql`
    mutation Infra_deleteProviderSecret($secretName: String!) {
      infra_deleteProviderSecret(secretName: $secretName)
    }
  `),
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
      dataPath: 'infra_getProviderSecret',
    }
  ),
});
