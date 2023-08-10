import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const providerSecretQueries = (
  executor = ExecuteQueryWithContext({})
) => ({
  listProviderSecrets: executor(
    gql`
      query Edges($pagination: PaginationQueryArgs, $search: SearchFilter) {
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
});
