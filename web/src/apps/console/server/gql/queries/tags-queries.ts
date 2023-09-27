import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
    ConsoleDeleteTagMutation,
    ConsoleDeleteTagMutationVariables,
    ConsoleListTagsQuery,
    ConsoleListTagsQueryVariables,
} from '~/root/src/generated/gql/server';

export type ITags = NN<ConsoleListTagsQuery['cr_listTags']>;

export const tagsQueries = (executor: IExecutor) => ({
  listTags: executor(
    gql`
      query Cr_listTags(
        $repoName: String!
        $search: SearchRepos
        $pagination: CursorPaginationIn
      ) {
        cr_listTags(
          repoName: $repoName
          search: $search
          pagination: $pagination
        ) {
          edges {
            cursor
            node {
              accountName
              actor
              creationTime
              deleting
              digest
              id
              length
              markedForDeletion
              mediaType
              recordVersion
              references {
                digest
                mediaType
                size
              }
              repository
              size
              tags
              updateTime
              url
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
      transformer(data: ConsoleListTagsQuery) {
        return data.cr_listTags;
      },
      vars(_: ConsoleListTagsQueryVariables) {},
    }
  ),
  deleteTag: executor(
    gql`
      mutation Cr_deleteTag($repoName: String!, $digest: String!) {
        cr_deleteTag(repoName: $repoName, digest: $digest)
      }
    `,
    {
      transformer(data: ConsoleDeleteTagMutation) {
        return data.cr_deleteTag;
      },
      vars(_: ConsoleDeleteTagMutationVariables) {},
    }
  ),
});
