import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleListPvcsQuery,
  ConsoleListPvcsQueryVariables,
} from '~/root/src/generated/gql/server';

export type IPvcs = NN<ConsoleListPvcsQuery['infra_listPVCs']>;

export const pvcQueries = (executor: IExecutor) => ({
  getPvc: executor(
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
  listPvcs: executor(
    gql`
      query Infra_listPVCs(
        $clusterName: String!
        $search: SearchPersistentVolumeClaims
        $pq: CursorPaginationIn
      ) {
        infra_listPVCs(clusterName: $clusterName, search: $search, pq: $pq) {
          edges {
            cursor
            node {
              creationTime
              id
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
                accessModes
                dataSource {
                  apiGroup
                  kind
                  name
                }
                dataSourceRef {
                  apiGroup
                  kind
                  name
                  namespace
                }
                resources {
                  claims {
                    name
                  }
                  limits
                  requests
                }
                selector {
                  matchExpressions {
                    key
                    operator
                    values
                  }
                  matchLabels
                }
                storageClassName
                volumeMode
                volumeName
              }
              status {
                accessModes
                allocatedResources
                allocatedResourceStatuses
                capacity
                conditions {
                  lastProbeTime
                  lastTransitionTime
                  message
                  reason
                  status
                  type
                }
                phase
              }
              updateTime
            }
          }
          pageInfo {
            endCursor
            hasNextPage
            hasPreviousPage
            startCursor
          }
          totalCount
        }
      }
    `,
    {
      transformer: (data: ConsoleListPvcsQuery) => data.infra_listPVCs,
      vars(_: ConsoleListPvcsQueryVariables) {},
    }
  ),
});
