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
              markedForDeletion
              metadata {
                name
                annotations
              }
              creationTime
              lastUpdatedBy {
                userId
                userName
                userEmail
              }
              createdBy {
                userEmail
                userId
                userName
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
                messageQueueTopicName
                kloudliteRelease

                credentialsRef {
                  namespace
                  name
                }
                clusterTokenRef {
                  key
                  name
                  namespace
                }
                accountId
                accountName
                availabilityMode
                aws {
                  k3sMasters {
                    iamInstanceProfileRole
                    instanceType
                    nodes
                    nvidiaGpuEnabled

                    rootVolumeSize
                    rootVolumeType
                  }
                  nodePools
                  region
                  spotNodePools
                }
                cloudProvider
                backupToS3Enabled
                cloudflareEnabled
                clusterInternalDnsHost
                output {
                  keyK3sAgentJoinToken
                  keyK3sServerJoinToken
                  keyKubeconfig
                  secretName
                }
                publicDNSHost
                taintMasterNodes
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
          accountName
          apiVersion
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
          spec {
            accountId
            accountName
            availabilityMode
            aws {
              k3sMasters {
                iamInstanceProfileRole
                imageId
                imageSSHUsername
                instanceType
                nodes
                nvidiaGpuEnabled
                rootVolumeSize
                rootVolumeType
              }
              nodePools
              region
              spotNodePools
            }
            backupToS3Enabled
            cloudflareEnabled
            cloudProvider
            clusterInternalDnsHost
            clusterTokenRef {
              key
              name
              namespace
            }
            credentialKeys {
              keyAccessKey
              keyAWSAccountId
              keyAWSAssumeRoleExternalID
              keyAWSAssumeRoleRoleARN
              keyIAMInstanceProfileRole
              keySecretKey
            }
            credentialsRef {
              name
              namespace
            }
            kloudliteRelease
            messageQueueTopicName
            output {
              keyK3sAgentJoinToken
              keyK3sServerJoinToken
              keyKubeconfig
              secretName
            }
            publicDNSHost
            taintMasterNodes
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
          clusterToken
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
