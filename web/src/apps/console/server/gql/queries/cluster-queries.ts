import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleClustersCountQuery,
  ConsoleClustersCountQueryVariables,
  ConsoleCreateClusterMutation,
  ConsoleCreateClusterMutationVariables,
  ConsoleGetClusterQuery,
  ConsoleGetClusterQueryVariables,
  ConsoleListClustersQuery,
  ConsoleListClustersQueryVariables,
} from '~/root/src/generated/gql/server';

export type ICluster = NN<ConsoleGetClusterQuery['infra_getCluster']>;
export type IClusters = NN<ConsoleListClustersQuery['infra_listClusters']>;

export const clusterQueries = (executor: IExecutor) => ({
  createCluster: executor(
    gql`
      mutation CreateCluster($cluster: ClusterIn!) {
        infra_createCluster(cluster: $cluster) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateClusterMutation) =>
        data.infra_createCluster,
      vars(_: ConsoleCreateClusterMutationVariables) {},
    }
  ),
  clustersCount: executor(
    gql`
      query Infra_clustersCount {
        infra_listClusters {
          totalCount
        }
      }
    `,
    {
      transformer: (data: ConsoleClustersCountQuery) => data.infra_listClusters,
      vars(_: ConsoleClustersCountQueryVariables) {},
    }
  ),

  listClusters: executor(
    gql`
      query Infra_listClusterss(
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
              displayName
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
      transformer: (data: ConsoleListClustersQuery) => data.infra_listClusters,
      vars(_: ConsoleListClustersQueryVariables) {},
    }
  ),
  getCluster: executor(
    gql`
      query Infra_getCluster($name: String!) {
        infra_getCluster(name: $name) {
          displayName
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
    `,
    {
      transformer: (data: ConsoleGetClusterQuery) => data.infra_getCluster,
      vars(_: ConsoleGetClusterQueryVariables) {},
    }
  ),
});
