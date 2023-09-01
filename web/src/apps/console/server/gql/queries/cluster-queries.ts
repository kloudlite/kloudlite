import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGqlReturn } from '~/root/lib/types/common';

export interface IGQLMethodsCluster {
  createCluster: (variables?: any) => IGqlReturn<any>;
  clustersCount: (variables?: any) => IGqlReturn<any>;
  listClusters: (variables?: any) => IGqlReturn<any>;
  getCluster: (variables?: any) => IGqlReturn<any>;
}

export const clusterQueries = (executor: IExecutor): IGQLMethodsCluster => ({
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
        $search: SearchCluster
        $pagination: CursorPaginationIn
      ) {
        infra_listClusters(search: $search, pagination: $pagination) {
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
            annotations
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
