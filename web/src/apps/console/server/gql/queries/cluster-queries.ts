import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleClustersCountQuery,
  ConsoleClustersCountQueryVariables,
  ConsoleCreateClusterMutation,
  ConsoleCreateClusterMutationVariables,
  ConsoleDeleteClusterMutation,
  ConsoleDeleteClusterMutationVariables,
  ConsoleGetClusterQuery,
  ConsoleGetClusterQueryVariables,
  ConsoleListClustersQuery,
  ConsoleListClustersQueryVariables,
  ConsoleUpdateClusterMutation,
  ConsoleUpdateClusterMutationVariables,
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
  deleteCluster: executor(
    gql`
      mutation Infra_deleteCluster($name: String!) {
        infra_deleteCluster(name: $name)
      }
    `,
    {
      transformer: (data: ConsoleDeleteClusterMutation) =>
        data.infra_deleteCluster,
      vars(_: ConsoleDeleteClusterMutationVariables) {},
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

                credentialsRef {
                  namespace
                  name
                }
                cloudProvider
                availabilityMode
                aws {
                  spotNodesConfig
                  spotSettings {
                    spotFleetTaggingRoleName
                  }
                  region
                  iamInstanceProfileRole
                  ec2NodesConfig
                  ami
                }
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
            accountId
            accountName
            agentHelmValuesRef {
              key
              name
              namespace
            }
            availabilityMode
            aws {
              ami
              ec2NodesConfig
              iamInstanceProfileRole
              region
              spotNodesConfig
              spotSettings {
                spotFleetTaggingRoleName
                enabled
              }
            }
            cloudProvider
            credentialsRef {
              name
              namespace
            }
            disableSSH
            nodeIps
            operatorsHelmValuesRef {
              key
              name
              namespace
            }
            vpc
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleGetClusterQuery) => data.infra_getCluster,
      vars(_: ConsoleGetClusterQueryVariables) {},
    }
  ),
  updateCluster: executor(
    gql`
      mutation Infra_updateCluster($cluster: ClusterIn!) {
        infra_updateCluster(cluster: $cluster) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateClusterMutation) =>
        data.infra_updateCluster,
      vars(_: ConsoleUpdateClusterMutationVariables) {},
    }
  ),
});
