import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleGetPvcQuery,
  ConsoleGetPvcQueryVariables,
  ConsoleListPvcsQuery,
  ConsoleListPvcsQueryVariables,
} from '~/root/src/generated/gql/server';

export type IPvcs = NN<ConsoleListPvcsQuery['infra_listPVCs']>;

export const pvcQueries = (executor: IExecutor) => ({
  getPvc: executor(
    gql`
      query Infra_getPVC($clusterName: String!, $name: String!) {
        infra_getPVC(clusterName: $clusterName, name: $name) {
          clusterName
          creationTime
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
    `,
    {
      transformer(data: ConsoleGetPvcQuery) {
        return data.infra_getPVC;
      },
      vars(_: ConsoleGetPvcQueryVariables) {},
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
