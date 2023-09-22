import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateCredMutation,
  ConsoleCreateCredMutationVariables,
  ConsoleDeleteCredMutation,
  ConsoleDeleteCredMutationVariables,
  ConsoleGetCredTokenQuery,
  ConsoleGetCredTokenQueryVariables,
  ConsoleListCredQuery,
  ConsoleListCredQueryVariables,
} from '~/root/src/generated/gql/server';

export type ICRCred = NN<ConsoleListCredQuery['cr_listCreds']>;

export const crQueries = (executor: IExecutor) => ({
  getCredToken: executor(
    gql`
      query Query($username: String!) {
        cr_getCredToken(username: $username)
      }
    `,
    {
      transformer: (data: ConsoleGetCredTokenQuery) => data.cr_getCredToken,
      vars(_: ConsoleGetCredTokenQueryVariables) {},
    }
  ),
  listCred: executor(
    gql`
      query Edges($search: SearchCreds, $pagination: CursorPaginationIn) {
        cr_listCreds(search: $search, pagination: $pagination) {
          edges {
            cursor
            node {
              access
              accountName
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              expiration {
                unit
                value
              }
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
              username
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
      transformer: (data: ConsoleListCredQuery) => data.cr_listCreds,
      vars(_: ConsoleListCredQueryVariables) {},
    }
  ),
  createCred: executor(
    gql`
      mutation Cr_createCred($credential: CredentialIn!) {
        cr_createCred(credential: $credential)
      }
    `,
    {
      transformer: (data: ConsoleCreateCredMutation) => data.cr_createCred,
      vars(_: ConsoleCreateCredMutationVariables) {},
    }
  ),
  deleteCred: executor(
    gql`
      mutation Cr_deleteCred($username: String!) {
        cr_deleteCred(username: $username)
      }
    `,
    {
      transformer: (data: ConsoleDeleteCredMutation) => data.cr_deleteCred,
      vars(_: ConsoleDeleteCredMutationVariables) {},
    }
  ),
});
