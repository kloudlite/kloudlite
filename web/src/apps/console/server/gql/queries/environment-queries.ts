import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateEnvironmentMutation,
  ConsoleCreateEnvironmentMutationVariables,
  ConsoleGetEnvironmentQuery,
  ConsoleGetEnvironmentQueryVariables,
  ConsoleListEnvironmentsQuery,
  ConsoleListEnvironmentsQueryVariables,
  ConsoleUpdateEnvironmentMutation,
  ConsoleUpdateEnvironmentMutationVariables,
  ConsoleDeleteEnvironmentMutation,
  ConsoleDeleteEnvironmentMutationVariables,
  ConsoleCloneEnvironmentMutation,
  ConsoleCloneEnvironmentMutationVariables,
} from '~/root/src/generated/gql/server';

export type IEnvironment = NN<
  ConsoleGetEnvironmentQuery['core_getEnvironment']
>;

export type IEnvironments = NN<
  ConsoleListEnvironmentsQuery['core_listEnvironments']
>;

export const environmentQueries = (executor: IExecutor) => ({
  getEnvironment: executor(
    gql`
      query Core_getEnvironment($projectName: String!, $name: String!) {
        core_getEnvironment(projectName: $projectName, name: $name) {
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
          projectName
          spec {
            projectName
            routing {
              mode
              privateIngressClass
              publicIngressClass
            }
            targetNamespace
          }
          status {
            checks
            checkList {
              description
              name
              title
            }
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
      transformer: (data: ConsoleGetEnvironmentQuery) =>
        data.core_getEnvironment,
      vars(_: ConsoleGetEnvironmentQueryVariables) {},
    }
  ),
  createEnvironment: executor(
    gql`
      mutation Core_createEnvironment(
        $projectName: String!
        $env: EnvironmentIn!
      ) {
        core_createEnvironment(projectName: $projectName, env: $env) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateEnvironmentMutation) =>
        data.core_createEnvironment,
      vars(_: ConsoleCreateEnvironmentMutationVariables) {},
    }
  ),
  updateEnvironment: executor(
    gql`
      mutation Core_updateEnvironment(
        $projectName: String!
        $env: EnvironmentIn!
      ) {
        core_updateEnvironment(projectName: $projectName, env: $env) {
          id
        }
      }
    `,
    {
      transformer(data: ConsoleUpdateEnvironmentMutation) {
        return data.core_updateEnvironment;
      },
      vars(_: ConsoleUpdateEnvironmentMutationVariables) {},
    }
  ),
  deleteEnvironment: executor(
    gql`
      mutation Core_deleteEnvironment(
        $projectName: String!
        $envName: String!
      ) {
        core_deleteEnvironment(projectName: $projectName, envName: $envName)
      }
    `,
    {
      transformer(data: ConsoleDeleteEnvironmentMutation) {
        return data.core_deleteEnvironment;
      },
      vars(_: ConsoleDeleteEnvironmentMutationVariables) {},
    }
  ),
  listEnvironments: executor(
    gql`
      query Core_listEnvironments(
        $projectName: String!
        $search: SearchEnvironments
        $pq: CursorPaginationIn
      ) {
        core_listEnvironments(
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
                generation
                name
                namespace
              }
              projectName
              recordVersion
              spec {
                projectName
                routing {
                  mode
                  privateIngressClass
                  publicIngressClass
                }
                targetNamespace
              }
              status {
                checks
                checkList {
                  description
                  name
                  title
                }
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
            hasPreviousPage
            startCursor
          }
          totalCount
        }
      }
    `,
    {
      transformer: (data: ConsoleListEnvironmentsQuery) =>
        data.core_listEnvironments,
      vars(_: ConsoleListEnvironmentsQueryVariables) {},
    }
  ),
  cloneEnvironment: executor(
    gql`
      mutation Core_cloneEnvironment(
        $projectName: String!
        $sourceEnvName: String!
        $destinationEnvName: String!
        $displayName: String!
        $environmentRoutingMode: Github__com___kloudlite___operator___apis___crds___v1__EnvironmentRoutingMode!
      ) {
        core_cloneEnvironment(
          projectName: $projectName
          sourceEnvName: $sourceEnvName
          destinationEnvName: $destinationEnvName
          displayName: $displayName
          environmentRoutingMode: $environmentRoutingMode
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCloneEnvironmentMutation) =>
        data.core_cloneEnvironment,
      vars(_: ConsoleCloneEnvironmentMutationVariables) {},
    }
  ),
});
