import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleListBuildRunsQuery,
  ConsoleListBuildRunsQueryVariables,
  ConsoleGetBuildRunQuery,
  ConsoleGetBuildRunQueryVariables,
} from '~/root/src/generated/gql/server';

export type IBuildRuns = NN<ConsoleListBuildRunsQuery['cr_listBuildRuns']>;

export const buildRunQueries = (executor: IExecutor) => ({
  listBuildRuns: executor(
    gql`
      query Cr_listBuildRuns(
        $repoName: String!
        $search: SearchBuildRuns
        $pq: CursorPaginationIn
      ) {
        cr_listBuildRuns(repoName: $repoName, search: $search, pq: $pq) {
          edges {
            cursor
            node {
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
                accountName
                buildOptions {
                  buildArgs
                  buildContexts
                  contextDir
                  dockerfileContent
                  dockerfilePath
                  targetPlatforms
                }
                cacheKeyName
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
      transformer: (data: ConsoleListBuildRunsQuery) => data.cr_listBuildRuns,
      vars(_: ConsoleListBuildRunsQueryVariables) {},
    }
  ),
  getBuildRun: executor(
    gql`
      query Cr_getBuildRun($repoName: String!, $buildRunName: String!) {
        cr_getBuildRun(repoName: $repoName, buildRunName: $buildRunName) {
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
            accountName
            buildOptions {
              buildArgs
              buildContexts
              contextDir
              dockerfileContent
              dockerfilePath
              targetPlatforms
            }
            cacheKeyName
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
      transformer: (data: ConsoleGetBuildRunQuery) => data.cr_getBuildRun,
      vars(_: ConsoleGetBuildRunQueryVariables) {},
    }
  ),
});
