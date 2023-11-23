import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleDeleteBuildCacheMutation,
  ConsoleDeleteBuildCacheMutationVariables,
  ConsoleUpdateBuildCachesMutation,
  ConsoleUpdateBuildCachesMutationVariables,
  ConsoleCreateBuildCacheMutation,
  ConsoleCreateBuildCacheMutationVariables,
  ConsoleListBuildCachesQuery,
  ConsoleListBuildCachesQueryVariables,
} from '~/root/src/generated/gql/server';

export type IBuildCaches = NN<
  ConsoleListBuildCachesQuery['cr_listBuildCacheKeys']
>;

export const buildCachesQueries = (executor: IExecutor) => ({
  listBuildCaches: executor(
    gql`
      query Cr_listBuildCacheKeys(
        $pq: CursorPaginationIn
        $search: SearchBuildCacheKeys
      ) {
        cr_listBuildCacheKeys(pq: $pq, search: $search) {
          edges {
            cursor
            node {
              id
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
              name
              updateTime
              volumeSizeInGB
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
      transformer: (data: ConsoleListBuildCachesQuery) =>
        data.cr_listBuildCacheKeys,
      vars(_: ConsoleListBuildCachesQueryVariables) {},
    }
  ),
  createBuildCache: executor(
    gql`
      mutation Cr_addBuildCacheKey($buildCacheKey: BuildCacheKeyIn!) {
        cr_addBuildCacheKey(buildCacheKey: $buildCacheKey) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateBuildCacheMutation) =>
        data.cr_addBuildCacheKey,
      vars(_: ConsoleCreateBuildCacheMutationVariables) {},
    }
  ),
  updateBuildCaches: executor(
    gql`
      mutation Cr_updateBuildCacheKey(
        $crUpdateBuildCacheKeyId: ID!
        $buildCacheKey: BuildCacheKeyIn!
      ) {
        cr_updateBuildCacheKey(
          id: $crUpdateBuildCacheKeyId
          buildCacheKey: $buildCacheKey
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateBuildCachesMutation) =>
        data.cr_updateBuildCacheKey,
      vars(_: ConsoleUpdateBuildCachesMutationVariables) {},
    }
  ),
  deleteBuildCache: executor(
    gql`
      mutation Cr_deleteBuildCacheKey($crDeleteBuildCacheKeyId: ID!) {
        cr_deleteBuildCacheKey(id: $crDeleteBuildCacheKeyId)
      }
    `,
    {
      transformer: (data: ConsoleDeleteBuildCacheMutation) =>
        data.cr_deleteBuildCacheKey,
      vars(_: ConsoleDeleteBuildCacheMutationVariables) {},
    }
  ),
});
