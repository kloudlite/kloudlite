import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleGetProjectMSvQuery,
  ConsoleListProjectMSvsQuery,
  ConsoleGetProjectMSvQueryVariables,
  ConsoleListProjectMSvsQueryVariables,
  ConsoleCreateProjectMSvMutation,
  ConsoleCreateProjectMSvMutationVariables,
  ConsoleUpdateProjectMSvMutation,
  ConsoleUpdateProjectMSvMutationVariables,
  ConsoleDeleteProjectMSvMutation,
  ConsoleDeleteProjectMSvMutationVariables,
} from '~/root/src/generated/gql/server';

export type IProjectMSv = NN<
  ConsoleGetProjectMSvQuery['core_getProjectManagedService']
>;
export type IProjectMSvs = NN<
  ConsoleListProjectMSvsQuery['core_listProjectManagedServices']
>;

export const projectManagedServicesQueries = (executor: IExecutor) => ({
  getProjectMSv: executor(
    gql`
      query Core_getProjectManagedService(
        $projectName: String!
        $name: String!
      ) {
        core_getProjectManagedService(projectName: $projectName, name: $name) {
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
          spec {
            msvcSpec {
              serviceTemplate {
                apiVersion
                kind
                spec
              }
            }
            targetNamespace
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
      transformer(data: ConsoleGetProjectMSvQuery) {
        return data.core_getProjectManagedService;
      },
      vars(_: ConsoleGetProjectMSvQueryVariables) { },
    }
  ),
  createProjectMSv: executor(
    gql`
      mutation Core_createProjectManagedService(
        $projectName: String!
        $pmsvc: ProjectManagedServiceIn!
      ) {
        core_createProjectManagedService(
          projectName: $projectName
          pmsvc: $pmsvc
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateProjectMSvMutation) =>
        data.core_createProjectManagedService,
      vars(_: ConsoleCreateProjectMSvMutationVariables) { },
    }
  ),
  updateProjectMSv: executor(
    gql`
      mutation Core_updateProjectManagedService(
        $projectName: String!
        $pmsvc: ProjectManagedServiceIn!
      ) {
        core_updateProjectManagedService(
          projectName: $projectName
          pmsvc: $pmsvc
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateProjectMSvMutation) =>
        data.core_updateProjectManagedService,
      vars(_: ConsoleUpdateProjectMSvMutationVariables) { },
    }
  ),
  listProjectMSvs: executor(
    gql`
      query Core_listProjectManagedServices(
        $projectName: String!
        $search: SearchProjectManagedService
        $pq: CursorPaginationIn
      ) {
        core_listProjectManagedServices(
          projectName: $projectName
          search: $search
          pq: $pq
        ) {
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
                targetNamespace
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
      transformer: (data: ConsoleListProjectMSvsQuery) =>
        data.core_listProjectManagedServices,
      vars(_: ConsoleListProjectMSvsQueryVariables) { },
    }
  ),
  deleteProjectMSv: executor(
    gql`
      mutation Core_deleteProjectManagedService(
        $projectName: String!
        $pmsvcName: String!
      ) {
        core_deleteProjectManagedService(
          projectName: $projectName
          pmsvcName: $pmsvcName
        )
      }
    `,
    {
      transformer: (data: ConsoleDeleteProjectMSvMutation) =>
        data.core_deleteProjectManagedService,
      vars(_: ConsoleDeleteProjectMSvMutationVariables) { },
    }
  ),
});
