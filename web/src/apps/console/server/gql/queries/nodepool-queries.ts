import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateNodePoolMutation,
  ConsoleCreateNodePoolMutationVariables,
  ConsoleGetNodePoolQuery,
  ConsoleGetNodePoolQueryVariables,
  ConsoleListNodePoolsQuery,
  ConsoleListNodePoolsQueryVariables,
  ConsoleDeleteNodePoolMutation,
  ConsoleDeleteNodePoolMutationVariables,
} from '~/root/src/generated/gql/server';

export type INodepool = NN<ConsoleGetNodePoolQuery['infra_getNodePool']>;
export type INodepools = NN<ConsoleListNodePoolsQuery['infra_listNodePools']>;

export const nodepoolQueries = (executor: IExecutor) => ({
  getNodePool: executor(
    gql`
      query Infra_getNodePool($clusterName: String!, $poolName: String!) {
        infra_getNodePool(clusterName: $clusterName, poolName: $poolName) {
          clusterName
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
          kind
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          markedForDeletion
          metadata {
            annotations
            creationTimestamp
            deletionTimestamp
            generation
            labels
            name
            namespace
          }
          spec {
            aws {
              availabilityZone
              ec2Pool {
                instanceType
                nodes
              }
              iamInstanceProfileRole
              imageId
              imageSSHUsername
              nvidiaGpuEnabled
              poolType
              rootVolumeSize
              rootVolumeType
              spotPool {
                cpuNode {
                  memoryPerVcpu {
                    max
                    min
                  }
                  vcpu {
                    max
                    min
                  }
                }
                gpuNode {
                  instanceTypes
                }
                nodes
                spotFleetTaggingRoleName
              }
            }
            cloudProvider
            maxCount
            minCount
            targetCount
          }
          status {
            checks
            isReady
            lastReadyGeneration
            lastReconcileTime
            message {
              RawMessage
            }
            resources {
              apiVersion
              kind
              name
              namespace
            }
          }
          syncStatus {
            action
            error
            lastSyncedAt
            recordVersion
            state
            syncScheduledAt
          }
          updateTime
        }
      }
    `,
    {
      transformer(data: ConsoleGetNodePoolQuery) {
        return data.infra_getNodePool;
      },
      vars(_: ConsoleGetNodePoolQueryVariables) {},
    }
  ),
  createNodePool: executor(
    gql`
      mutation Infra_createNodePool($clusterName: String!, $pool: NodePoolIn!) {
        infra_createNodePool(clusterName: $clusterName, pool: $pool) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateNodePoolMutation) =>
        data.infra_createNodePool,
      vars(_: ConsoleCreateNodePoolMutationVariables) {},
    }
  ),
  updateNodePool: executor(
    gql`
      mutation Infra_updateNodePool($clusterName: String!, $pool: NodePoolIn!) {
        infra_updateNodePool(clusterName: $clusterName, pool: $pool) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateNodePoolMutation) =>
        data.infra_createNodePool,
      vars(_: ConsoleCreateNodePoolMutationVariables) {},
    }
  ),
  listNodePools: executor(
    gql`
      query Infra_listNodePools(
        $clusterName: String!
        $search: SearchNodepool
        $pagination: CursorPaginationIn
      ) {
        infra_listNodePools(
          clusterName: $clusterName
          search: $search
          pagination: $pagination
        ) {
          totalCount
          pageInfo {
            endCursor
            hasNextPage
            startCursor
          }
          edges {
            cursor
            node {
              clusterName
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              markedForDeletion
              metadata {
                name
              }
              spec {
                aws {
                  availabilityZone
                  ec2Pool {
                    instanceType
                    nodes
                  }
                  nvidiaGpuEnabled
                  poolType
                  spotPool {
                    cpuNode {
                      memoryPerVcpu {
                        max
                        min
                      }
                      vcpu {
                        max
                        min
                      }
                    }
                    gpuNode {
                      instanceTypes
                    }
                    nodes
                  }
                }
                cloudProvider
                maxCount
                minCount
                targetCount
              }
              status {
                checks
                isReady
                lastReadyGeneration
                lastReconcileTime
                message {
                  RawMessage
                }
                resources {
                  apiVersion
                  kind
                  name
                  namespace
                }
              }
              syncStatus {
                action
                error
                lastSyncedAt
                recordVersion
                state
                syncScheduledAt
              }
              updateTime
            }
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleListNodePoolsQuery) =>
        data.infra_listNodePools,
      vars(_: ConsoleListNodePoolsQueryVariables) {},
    }
  ),
  deleteNodePool: executor(
    gql`
      mutation Infra_deleteNodePool($clusterName: String!, $poolName: String!) {
        infra_deleteNodePool(clusterName: $clusterName, poolName: $poolName)
      }
    `,
    {
      transformer: (data: ConsoleDeleteNodePoolMutation) =>
        data.infra_deleteNodePool,
      vars(_: ConsoleDeleteNodePoolMutationVariables) {},
    }
  ),
});
