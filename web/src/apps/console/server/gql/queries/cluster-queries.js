import { execute } from 'graphql';
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
  clustersCount: executor(
    gql`
      query Infra_listClusters {
        infra_listClusters {
          totalCount
        }
      }
    `,
    {
      dataPath: 'infra_listClusters',
    }
  ),

  listClusters: executor(
    gql`
      query Infra_listClusters(
        $pagination: PaginationQueryArgs
        $search: SearchFilter
      ) {
        infra_listClusters(pagination: $pagination, search: $search) {
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
              metadata {
                name
                annotations
              }
              updateTime
              syncStatus {
                syncScheduledAt
                lastSyncedAt
                recordVersion
                state
                error
                action
              }
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
              recordVersion
              spec {
                vpc
                region
                credentialsRef {
                  namespace
                  name
                }
                cloudProvider
                availabilityMode
              }
            }
          }
        }
      }
    `,
    {
      dataPath: 'infra_listClusters',
    }
  ),
  getCluster: executor(
    gql`
      query Infra_getCluster($name: String!) {
        infra_getCluster(name: $name) {
          metadata {
            name
          }
          spec {
            vpc
            region
            nodeIps
            cloudProvider
            availabilityMode
          }
        }
      }
    `,
    {
      dataPath: 'infra_getCluster',
    }
  ),
});
