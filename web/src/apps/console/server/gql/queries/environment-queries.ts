import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCloneEnvironmentMutation,
  ConsoleCloneEnvironmentMutationVariables,
  ConsoleCreateEnvironmentMutation,
  ConsoleCreateEnvironmentMutationVariables,
  ConsoleDeleteEnvironmentMutation,
  ConsoleDeleteEnvironmentMutationVariables,
  ConsoleGetEnvironmentQuery,
  ConsoleGetEnvironmentQueryVariables,
  ConsoleListEnvironmentsQuery,
  ConsoleListEnvironmentsQueryVariables,
  ConsoleSetupDefaultEnvironmentMutation,
  ConsoleSetupDefaultEnvironmentMutationVariables,
  ConsoleUpdateEnvironmentMutation,
  ConsoleUpdateEnvironmentMutationVariables,
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
      query Core_getEnvironment($name: String!) {
        core_getEnvironment(name: $name) {
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
          clusterName
          isArchived
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          markedForDeletion
          clusterName
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
            routing {
              mode
              privateIngressClass
              publicIngressClass
            }
            suspend
            targetNamespace
          }
          status {
            checks
            checkList {
              description
              debug
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
      vars(_: ConsoleGetEnvironmentQueryVariables) { },
    }
  ),
  createEnvironment: executor(
    gql`
      mutation Core_createEnvironment($env: EnvironmentIn!) {
        core_createEnvironment(env: $env) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateEnvironmentMutation) =>
        data.core_createEnvironment,
      vars(_: ConsoleCreateEnvironmentMutationVariables) { },
    }
  ),
  setupDefaultEnvironment: executor(
    gql`
      mutation Mutation {
        core_setupDefaultEnvironment
      }
    `,
    {
      transformer: (data: ConsoleSetupDefaultEnvironmentMutation) =>
        data.core_setupDefaultEnvironment,
      vars(_: ConsoleSetupDefaultEnvironmentMutationVariables) { },
    }
  ),
  updateEnvironment: executor(
    gql`
      mutation Core_updateEnvironment($env: EnvironmentIn!) {
        core_updateEnvironment(env: $env) {
          id
        }
      }
    `,
    {
      transformer(data: ConsoleUpdateEnvironmentMutation) {
        return data.core_updateEnvironment;
      },
      vars(_: ConsoleUpdateEnvironmentMutationVariables) { },
    }
  ),
  deleteEnvironment: executor(
    gql`
      mutation Core_deleteEnvironment($envName: String!) {
        core_deleteEnvironment(envName: $envName)
      }
    `,
    {
      transformer(data: ConsoleDeleteEnvironmentMutation) {
        return data.core_deleteEnvironment;
      },
      vars(_: ConsoleDeleteEnvironmentMutationVariables) { },
    }
  ),
  listEnvironments: executor(
    gql`
      query Core_listEnvironments(
        $search: SearchEnvironments
        $pq: CursorPaginationIn
      ) {
        core_listEnvironments(search: $search, pq: $pq) {
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
              clusterName
              isArchived
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
              recordVersion
              spec {
                routing {
                  mode
                  privateIngressClass
                  publicIngressClass
                }
                suspend
                targetNamespace
              }
              status {
                checks
                checkList {
                  description
                  debug
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
            hasPrevPage
            startCursor
          }
          totalCount
        }
      }
    `,
    {
      transformer: (data: ConsoleListEnvironmentsQuery) =>
        data.core_listEnvironments,
      vars(_: ConsoleListEnvironmentsQueryVariables) { },
    }
  ),
  cloneEnvironment: executor(
    gql`
      mutation Core_cloneEnvironment(
        $clusterName: String!
        $sourceEnvName: String!
        $destinationEnvName: String!
        $displayName: String!
        $environmentRoutingMode: Github__com___kloudlite___operator___apis___crds___v1__EnvironmentRoutingMode!
      ) {
        core_cloneEnvironment(
          clusterName: $clusterName
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
      vars(_: ConsoleCloneEnvironmentMutationVariables) { },
    }
  ),
});
