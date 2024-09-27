import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleGetPvQuery,
  ConsoleGetPvQueryVariables,
  ConsoleListPvsQuery,
  ConsoleListPvsQueryVariables,
  ConsoleDeletePvMutationVariables,
  ConsoleDeletePvMutation,
} from '~/root/src/generated/gql/server';

export type IPvs = NN<ConsoleListPvsQuery['infra_listPVs']>;

export const pvQueries = (executor: IExecutor) => ({
  getPv: executor(
    gql`
      query Infra_getPV($clusterName: String!, $name: String!) {
        infra_getPV(clusterName: $clusterName, name: $name) {
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
            accessModes
            awsElasticBlockStore {
              fsType
              partition
              readOnly
              volumeID
            }
            azureDisk {
              cachingMode
              diskName
              diskURI
              fsType
              kind
              readOnly
            }
            azureFile {
              readOnly
              secretName
              secretNamespace
              shareName
            }
            capacity
            cephfs {
              monitors
              path
              readOnly
              secretFile
              secretRef {
                name
                namespace
              }
              user
            }
            cinder {
              fsType
              readOnly
              volumeID
            }
            claimRef {
              apiVersion
              fieldPath
              kind
              name
              namespace
              resourceVersion
              uid
            }
            csi {
              controllerExpandSecretRef {
                name
                namespace
              }
              controllerPublishSecretRef {
                name
                namespace
              }
              driver
              fsType
              nodeExpandSecretRef {
                name
                namespace
              }
              nodePublishSecretRef {
                name
                namespace
              }
              nodeStageSecretRef {
                name
                namespace
              }
              readOnly
              volumeAttributes
              volumeHandle
            }
            fc {
              fsType
              lun
              readOnly
              targetWWNs
              wwids
            }
            flexVolume {
              driver
              fsType
              options
              readOnly
            }
            flocker {
              datasetName
              datasetUUID
            }
            gcePersistentDisk {
              fsType
              partition
              pdName
              readOnly
            }
            glusterfs {
              endpoints
              endpointsNamespace
              path
              readOnly
            }
            hostPath {
              path
              type
            }
            iscsi {
              chapAuthDiscovery
              chapAuthSession
              fsType
              initiatorName
              iqn
              iscsiInterface
              lun
              portals
              readOnly
              targetPortal
            }
            local {
              fsType
              path
            }
            mountOptions
            nfs {
              path
              readOnly
              server
            }
            nodeAffinity {
              required {
                nodeSelectorTerms {
                  matchExpressions {
                    key
                    operator
                    values
                  }
                  matchFields {
                    key
                    operator
                    values
                  }
                }
              }
            }
            persistentVolumeReclaimPolicy
            photonPersistentDisk {
              fsType
              pdID
            }
            portworxVolume {
              fsType
              readOnly
              volumeID
            }
            quobyte {
              group
              readOnly
              registry
              tenant
              user
              volume
            }
            rbd {
              fsType
              image
              keyring
              monitors
              pool
              readOnly
              user
            }
            scaleIO {
              fsType
              gateway
              protectionDomain
              readOnly
              sslEnabled
              storageMode
              storagePool
              system
              volumeName
            }
            storageClassName
            storageos {
              fsType
              readOnly
              secretRef {
                apiVersion
                fieldPath
                kind
                name
                namespace
                resourceVersion
                uid
              }
              volumeName
              volumeNamespace
            }
            volumeMode
            vsphereVolume {
              fsType
              storagePolicyID
              storagePolicyName
              volumePath
            }
          }
          status {
            lastPhaseTransitionTime
            message
            phase
            reason
          }
          updateTime
        }
      }
    `,
    {
      transformer(data: ConsoleGetPvQuery) {
        return data.infra_getPV;
      },
      vars(_: ConsoleGetPvQueryVariables) { },
    }
  ),
  listPvs: executor(
    gql`
      query Infra_listPVs(
        $clusterName: String!
        $search: SearchPersistentVolumes
        $pq: CursorPaginationIn
      ) {
        infra_listPVs(clusterName: $clusterName, search: $search, pq: $pq) {
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
                accessModes
                awsElasticBlockStore {
                  fsType
                  partition
                  readOnly
                  volumeID
                }
                azureDisk {
                  cachingMode
                  diskName
                  diskURI
                  fsType
                  kind
                  readOnly
                }
                azureFile {
                  readOnly
                  secretName
                  secretNamespace
                  shareName
                }
                capacity
                cephfs {
                  monitors
                  path
                  readOnly
                  secretFile
                  secretRef {
                    name
                    namespace
                  }
                  user
                }
                cinder {
                  fsType
                  readOnly
                  volumeID
                }
                claimRef {
                  apiVersion
                  fieldPath
                  kind
                  name
                  namespace
                  resourceVersion
                  uid
                }
                csi {
                  controllerExpandSecretRef {
                    name
                    namespace
                  }
                  controllerPublishSecretRef {
                    name
                    namespace
                  }
                  driver
                  fsType
                  nodeExpandSecretRef {
                    name
                    namespace
                  }
                  nodePublishSecretRef {
                    name
                    namespace
                  }
                  nodeStageSecretRef {
                    name
                    namespace
                  }
                  readOnly
                  volumeAttributes
                  volumeHandle
                }
                fc {
                  fsType
                  lun
                  readOnly
                  targetWWNs
                  wwids
                }
                flexVolume {
                  driver
                  fsType
                  options
                  readOnly
                }
                flocker {
                  datasetName
                  datasetUUID
                }
                gcePersistentDisk {
                  fsType
                  partition
                  pdName
                  readOnly
                }
                glusterfs {
                  endpoints
                  endpointsNamespace
                  path
                  readOnly
                }
                hostPath {
                  path
                  type
                }
                iscsi {
                  chapAuthDiscovery
                  chapAuthSession
                  fsType
                  initiatorName
                  iqn
                  iscsiInterface
                  lun
                  portals
                  readOnly
                  targetPortal
                }
                local {
                  fsType
                  path
                }
                mountOptions
                nfs {
                  path
                  readOnly
                  server
                }

                persistentVolumeReclaimPolicy
                photonPersistentDisk {
                  fsType
                  pdID
                }
                portworxVolume {
                  fsType
                  readOnly
                  volumeID
                }
                quobyte {
                  group
                  readOnly
                  registry
                  tenant
                  user
                  volume
                }
                rbd {
                  fsType
                  image
                  keyring
                  monitors
                  pool
                  readOnly
                  user
                }
                scaleIO {
                  fsType
                  gateway
                  protectionDomain
                  readOnly
                  sslEnabled
                  storageMode
                  storagePool
                  system
                  volumeName
                }
                storageClassName
                storageos {
                  fsType
                  readOnly
                  secretRef {
                    apiVersion
                    fieldPath
                    kind
                    name
                    namespace
                    resourceVersion
                    uid
                  }
                  volumeName
                  volumeNamespace
                }
                volumeMode
                vsphereVolume {
                  fsType
                  storagePolicyID
                  storagePolicyName
                  volumePath
                }
              }
              status {
                lastPhaseTransitionTime
                message
                phase
                reason
              }
              updateTime
            }
          }
          pageInfo {
            endCursor
            hasNextPage
            hasPrevPage
            startCursor
          }
          totalCount
        }
      }
    `,
    {
      transformer: (data: ConsoleListPvsQuery) => data.infra_listPVs,
      vars(_: ConsoleListPvsQueryVariables) { },
    }
  ),
  deletePV: executor(
    gql`
      mutation Infra_deletePV($clusterName: String!, $pvName: String!) {
        infra_deletePV(clusterName: $clusterName, pvName: $pvName)
      }
    `,
    {
      transformer: (data: ConsoleDeletePvMutation) => data.infra_deletePV,
      vars(_: ConsoleDeletePvMutationVariables) { },
    }
  ),
});
