import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleDeleteDigestMutation,
  ConsoleDeleteDigestMutationVariables,
  ConsoleListDigestQuery,
  ConsoleListDigestQueryVariables,
} from '~/root/src/generated/gql/server';

export type IDigests = NN<ConsoleListDigestQuery['cr_listDigests']>;

export const tagsQueries = (executor: IExecutor) => ({
  listDigest: executor(
    gql`
      query Cr_listDigests(
        $repoName: String!
        $search: SearchRepos
        $pagination: CursorPaginationIn
      ) {
        cr_listDigests(
          repoName: $repoName
          search: $search
          pagination: $pagination
        ) {
          pageInfo {
            endCursor
            hasNextPage
            hasPreviousPage
            startCursor
          }
          totalCount
          edges {
            cursor
            node {
              url
              updateTime
              tags
              size
              repository
              digest
              creationTime
            }
          }
        }
      }
    `,
    {
      transformer(data: ConsoleListDigestQuery) {
        return data.cr_listDigests;
      },
      vars(_: ConsoleListDigestQueryVariables) {},
    }
  ),
  deleteDigest: executor(
    gql`
      mutation Cr_deleteDigest($repoName: String!, $digest: String!) {
        cr_deleteDigest(repoName: $repoName, digest: $digest)
      }
    `,
    {
      transformer(data: ConsoleDeleteDigestMutation) {
        return data.cr_deleteDigest;
      },
      vars(_: ConsoleDeleteDigestMutationVariables) {},
    }
  ),
});
