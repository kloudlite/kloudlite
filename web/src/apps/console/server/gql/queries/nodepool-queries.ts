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
} from '~/root/src/generated/gql/server';

export type INodepool = NN<ConsoleGetNodePoolQuery['infra_getNodePool']>;
export type INodepools = NN<ConsoleListNodePoolsQuery['infra_listNodePools']>;

export const nodepoolQueries = (executor: IExecutor) => ({
  getNodePool: executor(
    gql`
      query User($clusterName: String!, $poolName: String!) {
        infra_getNodePool(clusterName: $clusterName, poolName: $poolName) {
          updateTime
          spec {
            targetCount
            minCount
            maxCount
            cloudProvider
            aws {
              normalPool {
                ami
                amiSSHUsername
                availabilityZone
                iamInstanceProfileRole
                instanceType
                nodes
                nvidiaGpuEnabled
                rootVolumeSize
                rootVolumeType
              }
              poolType
              spotPool {
                ami
                amiSSHUsername
                availabilityZone
                cpuNode {
                  vcpu {
                    max
                    min
                  }
                  memoryPerVcpu {
                    max
                    min
                  }
                }
                gpuNode {
                  instanceTypes
                }
                iamInstanceProfileRole
                nodes
                nvidiaGpuEnabled
                rootVolumeSize
                rootVolumeType
                spotFleetTaggingRoleName
              }
            }
          }
          metadata {
            name
            annotations
          }
          clusterName
          status {
            isReady
            message {
              RawMessage
            }
            checks
          }
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
  listNodePools: executor(
    gql`
      query Aws($clusterName: String!) {
        infra_listNodePools(clusterName: $clusterName) {
          pageInfo {
            hasNextPage
            hasPreviousPage
            startCursor
            endCursor
          }
          totalCount
          edges {
            node {
              spec {
                aws {
                  normalPool {
                    ami
                    amiSSHUsername
                    availabilityZone
                    iamInstanceProfileRole
                    instanceType
                    nodes
                    nvidiaGpuEnabled
                    rootVolumeSize
                    rootVolumeType
                  }
                  poolType
                  spotPool {
                    ami
                    amiSSHUsername
                    availabilityZone
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
                    iamInstanceProfileRole
                    nodes
                    nvidiaGpuEnabled
                    rootVolumeSize
                    rootVolumeType
                    spotFleetTaggingRoleName
                  }
                }
              }
              accountName
              apiVersion
              clusterName
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              id
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
              recordVersion
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
});
