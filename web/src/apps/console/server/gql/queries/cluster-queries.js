import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const clusterQueries = (executor = ExecuteQueryWithContext({})) => ({
  createCluster: executor(gql`
    mutation Mutation($cluster: ClusterIn!) {
      infra_createCluster(cluster: $cluster) {
        id
      }
    }
  `),
  listClusters: executor(
    gql`
      query Infra_listClusters($pagination: PaginationQueryArgs) {
        infra_listClusters(pagination: $pagination) {
          totalCount
          pageInfo {
            startCursor
            hasPreviousPage
            hasNextPage
            endCursor
          }
          edges {
            cursor
            node {
              accountName
              apiVersion
              creationTime
              id
              kind
              metadata {
                namespace
                name
                labels
                generation
                deletionTimestamp
                creationTimestamp
                annotations
              }
              recordVersion
              spec {
                region
                providerName
                provider
                count
                config
                accountName
              }
              status {
                resources {
                  kind
                  apiVersion
                  name
                  namespace
                }
                message {
                  RawMessage
                }
                lastReconcileTime
                isReady
                checks
              }
              syncStatus {
                syncScheduledAt
                state
                recordVersion
                lastSyncedAt
                error
                action
              }
              updateTime
            }
          }
        }
      }
    `,
    {
      dataPath: 'infra_listClusters',
    }
  ),
});
