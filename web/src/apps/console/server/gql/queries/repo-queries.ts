import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
    ConsoleCreateRepoMutation,
    ConsoleCreateRepoMutationVariables,
    ConsoleDeleteRepoMutation,
    ConsoleDeleteRepoMutationVariables,
    ConsoleListRepoQuery,
    ConsoleListRepoQueryVariables,
} from '~/root/src/generated/gql/server';

export type IRepos = NN<ConsoleListRepoQuery['cr_listRepos']>;

export const repoQueries = (executor: IExecutor) => ({
  listRepo: executor(
    gql`
      query Cr_listRepos(
        $search: SearchRepos
        $pagination: CursorPaginationIn
      ) {
        cr_listRepos(search: $search, pagination: $pagination) {
          edges {
            cursor
            node {
              accountName
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              id
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              markedForDeletion
              name
              recordVersion
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
      transformer: (data: ConsoleListRepoQuery) => data.cr_listRepos,
      vars(_: ConsoleListRepoQueryVariables) {},
    }
  ),
  createRepo: executor(
    gql`
      mutation Cr_createRepo($repository: RepositoryIn!) {
        cr_createRepo(repository: $repository)
      }
    `,
    {
      transformer: (data: ConsoleCreateRepoMutation) => data.cr_createRepo,
      vars(_: ConsoleCreateRepoMutationVariables) {},
    }
  ),
  deleteRepo: executor(
    gql`
      mutation Cr_deleteRepo($name: String!) {
        cr_deleteRepo(name: $name)
      }
    `,
    {
      transformer: (data: ConsoleDeleteRepoMutation) => data.cr_deleteRepo,
      vars(_: ConsoleDeleteRepoMutationVariables) {},
    }
  ),
});
