import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleGetClusterMSvQuery,
  ConsoleGetClusterMSvQueryVariables,
  ConsoleListClusterMSvsQuery,
  ConsoleListClusterMSvsQueryVariables,
  ConsoleCreateClusterMSvMutation,
  ConsoleCreateClusterMSvMutationVariables,
  ConsoleUpdateClusterMSvMutation,
  ConsoleUpdateClusterMSvMutationVariables,
  ConsoleDeleteClusterMSvMutation,
  ConsoleDeleteClusterMSvMutationVariables,
  ConsoleCloneClusterMSvMutation,
  ConsoleCloneClusterMSvMutationVariables,
} from '~/root/src/generated/gql/server';

export type IClusterMSv = NN<
  ConsoleGetClusterMSvQuery['infra_getClusterManagedService']
>;
export type IClusterMSvs = NN<
  ConsoleListClusterMSvsQuery['infra_listClusterManagedServices']
>;

export type InClusterMSvs = ConsoleUpdateClusterMSvMutationVariables;

export const clusterManagedServicesQueries = (executor: IExecutor) => ({
  getClusterMSv: executor(
    gql`
      query Infra_getClusterManagedService($name: String!) {
        infra_getClusterManagedService(name: $name) {
          clusterName
          creationTime
          displayName
          isArchived
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
            msvcSpec {
              nodeSelector
              serviceTemplate {
                apiVersion
                kind
                spec
              }
              tolerations {
                effect
                key
                operator
                tolerationSeconds
                value
              }
            }
            targetNamespace
          }
          updateTime
        }
      }
    `,
    {
      transformer(data: ConsoleGetClusterMSvQuery) {
        return data.infra_getClusterManagedService;
      },
      vars(_: ConsoleGetClusterMSvQueryVariables) {},
    }
  ),
  createClusterMSv: executor(
    gql`
      mutation Infra_createClusterManagedService(
        $service: ClusterManagedServiceIn!
      ) {
        infra_createClusterManagedService(service: $service) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateClusterMSvMutation) =>
        data.infra_createClusterManagedService,
      vars(_: ConsoleCreateClusterMSvMutationVariables) {},
    }
  ),
  cloneClusterMSv: executor(
    gql`
      mutation Infra_cloneClusterManagedService(
        $clusterName: String!
        $sourceMsvcName: String!
        $destinationMsvcName: String!
        $displayName: String!
      ) {
        infra_cloneClusterManagedService(
          clusterName: $clusterName
          sourceMsvcName: $sourceMsvcName
          destinationMsvcName: $destinationMsvcName
          displayName: $displayName
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCloneClusterMSvMutation) =>
        data.infra_cloneClusterManagedService,
      vars(_: ConsoleCloneClusterMSvMutationVariables) {},
    }
  ),
  updateClusterMSv: executor(
    gql`
      mutation Infra_updateClusterManagedService(
        $service: ClusterManagedServiceIn!
      ) {
        infra_updateClusterManagedService(service: $service) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateClusterMSvMutation) =>
        data.infra_updateClusterManagedService,
      vars(_: ConsoleUpdateClusterMSvMutationVariables) {},
    }
  ),
  listClusterMSvs: executor(
    gql`
      query Infra_listClusterManagedServices(
        $pagination: CursorPaginationIn
        $search: SearchClusterManagedService
      ) {
        infra_listClusterManagedServices(
          pagination: $pagination
          search: $search
        ) {
          edges {
            cursor
            node {
              accountName
              apiVersion
              clusterName
              isArchived
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
                msvcSpec {
                  nodeSelector
                  serviceTemplate {
                    apiVersion
                    kind
                    spec
                  }
                  tolerations {
                    effect
                    key
                    operator
                    tolerationSeconds
                    value
                  }
                }
                targetNamespace
              }
              status {
                checkList {
                  debug
                  description
                  hide
                  name
                  title
                }
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
      transformer: (data: ConsoleListClusterMSvsQuery) =>
        data.infra_listClusterManagedServices,
      vars(_: ConsoleListClusterMSvsQueryVariables) {},
    }
  ),
  deleteClusterMSv: executor(
    gql`
      mutation Infra_deleteClusterManagedService($name: String!) {
        infra_deleteClusterManagedService(name: $name)
      }
    `,
    {
      transformer: (data: ConsoleDeleteClusterMSvMutation) =>
        data.infra_deleteClusterManagedService,
      vars(_: ConsoleDeleteClusterMSvMutationVariables) {},
    }
  ),
});
