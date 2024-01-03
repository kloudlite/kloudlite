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
} from '~/root/src/generated/gql/server';

export type IClusterMSv = NN<
  ConsoleGetClusterMSvQuery['infra_getClusterManagedService']
>;
export type IClusterMSvs = NN<
  ConsoleListClusterMSvsQuery['infra_listClusterManagedServices']
>;

export const clusterManagedServicesQueries = (executor: IExecutor) => ({
  getClusterMSv: executor(
    gql`
      query Infra_getClusterManagedService(
        $clusterName: String!
        $name: String!
      ) {
        infra_getClusterManagedService(clusterName: $clusterName, name: $name) {
          displayName
          creationTime
          createdBy {
            userEmail
            userId
            userName
          }
          clusterName
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
            msvcSpec {
              serviceTemplate {
                apiVersion
                kind
                spec
              }
            }
            namespace
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
          updateTime
        }
      }
    `,
    {
      transformer(data: ConsoleGetClusterMSvQuery) {
        return data.infra_getClusterManagedService;
      },
      vars(_: ConsoleGetClusterMSvQueryVariables) { },
    }
  ),
  createClusterMSv: executor(
    gql`
      mutation Infra_createClusterManagedService(
        $clusterName: String!
        $service: ClusterManagedServiceIn!
      ) {
        infra_createClusterManagedService(
          clusterName: $clusterName
          service: $service
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateClusterMSvMutation) =>
        data.infra_createClusterManagedService,
      vars(_: ConsoleCreateClusterMSvMutationVariables) { },
    }
  ),
  updateClusterMSv: executor(
    gql`
      mutation Infra_updateClusterManagedService(
        $clusterName: String!
        $service: ClusterManagedServiceIn!
      ) {
        infra_updateClusterManagedService(
          clusterName: $clusterName
          service: $service
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateClusterMSvMutation) =>
        data.infra_updateClusterManagedService,
      vars(_: ConsoleUpdateClusterMSvMutationVariables) { },
    }
  ),
  listClusterMSvs: executor(
    gql`
      query Infra_listClusterManagedServices($clusterName: String!) {
        infra_listClusterManagedServices(clusterName: $clusterName) {
          edges {
            cursor
            node {
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
                msvcSpec {
                  serviceTemplate {
                    apiVersion
                    kind
                    spec
                  }
                }
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
      transformer: (data: ConsoleListClusterMSvsQuery) =>
        data.infra_listClusterManagedServices,
      vars(_: ConsoleListClusterMSvsQueryVariables) { },
    }
  ),
  deleteClusterMSv: executor(
    gql`
      mutation Infra_deleteClusterManagedService(
        $clusterName: String!
        $serviceName: String!
      ) {
        infra_deleteClusterManagedService(
          clusterName: $clusterName
          serviceName: $serviceName
        )
      }
    `,
    {
      transformer: (data: ConsoleDeleteClusterMSvMutation) =>
        data.infra_deleteClusterManagedService,
      vars(_: ConsoleDeleteClusterMSvMutationVariables) { },
    }
  ),
});
