import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateBuildMutation,
  ConsoleCreateBuildMutationVariables,
  ConsoleDeleteBuildMutation,
  ConsoleDeleteBuildMutationVariables,
  ConsoleListBuildsQuery,
  ConsoleListBuildsQueryVariables,
  ConsoleTriggerBuildMutation,
  ConsoleTriggerBuildMutationVariables,
  ConsoleUpdateBuildMutation,
  ConsoleUpdateBuildMutationVariables,
} from '~/root/src/generated/gql/server';

export type IBuilds = NN<ConsoleListBuildsQuery['cr_listBuilds']>;

export const iotBuildQueries = (executor: IExecutor) => ({
  listBuilds: executor(
    gql`
      query Cr_listBuilds(
        $repoName: String!
        $search: SearchBuilds
        $pagination: CursorPaginationIn
      ) {
        cr_listBuilds(
          repoName: $repoName
          search: $search
          pagination: $pagination
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
              buildClusterName
              credUser {
                userEmail
                userId
                userName
              }
              errorMessages
              id
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              markedForDeletion
              name
              source {
                branch
                provider
                repository
                webhookId
              }
              spec {
                buildOptions {
                  buildArgs
                  buildContexts
                  contextDir
                  dockerfileContent
                  dockerfilePath
                  targetPlatforms
                }
                registry {
                  repo {
                    name
                    tags
                  }
                }
                resource {
                  cpu
                  memoryInMb
                }
                caches {
                  name
                  path
                }
              }
              status
              updateTime
              latestBuildRun {
                recordVersion
                status {
                  checkList {
                    debug
                    description
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
                markedForDeletion
              }
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
      transformer: (data: ConsoleListBuildsQuery) => data.cr_listBuilds,
      vars(_: ConsoleListBuildsQueryVariables) {},
    }
  ),
  createBuild: executor(
    gql`
      mutation Cr_addBuild($build: BuildIn!) {
        cr_addBuild(build: $build) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateBuildMutation) => data.cr_addBuild,
      vars(_: ConsoleCreateBuildMutationVariables) {},
    }
  ),
  updateBuild: executor(
    gql`
      mutation Cr_updateBuild($crUpdateBuildId: ID!, $build: BuildIn!) {
        cr_updateBuild(id: $crUpdateBuildId, build: $build) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateBuildMutation) => data.cr_updateBuild,
      vars(_: ConsoleUpdateBuildMutationVariables) {},
    }
  ),
  deleteBuild: executor(
    gql`
      mutation Cr_deleteBuild($crDeleteBuildId: ID!) {
        cr_deleteBuild(id: $crDeleteBuildId)
      }
    `,
    {
      transformer: (data: ConsoleDeleteBuildMutation) => data.cr_deleteBuild,
      vars(_: ConsoleDeleteBuildMutationVariables) {},
    }
  ),
  triggerBuild: executor(
    gql`
      mutation Cr_triggerBuild($crTriggerBuildId: ID!) {
        cr_triggerBuild(id: $crTriggerBuildId)
      }
    `,
    {
      transformer: (data: ConsoleTriggerBuildMutation) => data.cr_triggerBuild,
      vars(_: ConsoleTriggerBuildMutationVariables) {},
    }
  ),
});
