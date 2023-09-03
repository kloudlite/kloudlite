import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  ConsoleCreateAccountMutation,
  ConsoleCreateAccountMutationVariables,
  ConsoleGetAccountQuery,
  ConsoleGetAccountQueryVariables,
  ConsoleListAccountsQuery,
  ConsoleListAccountsQueryVariables,
} from '~/root/src/generated/gql/server';
import { NN } from '~/root/src/generated/r-types/utils';

export type Accounts = NN<ConsoleListAccountsQuery['accounts_listAccounts']>;
export type Account = NN<ConsoleGetAccountQuery['accounts_getAccount']>;

export const accountQueries = (executor: IExecutor) => ({
  createAccount: executor(
    gql`
      mutation Accounts_createAccount($account: AccountIn!) {
        accounts_createAccount(account: $account) {
          displayName
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateAccountMutation) =>
        data.accounts_createAccount,
      vars(_: ConsoleCreateAccountMutationVariables) {},
    }
  ),

  listAccounts: executor(
    gql`
      query Accounts_listAccounts {
        accounts_listAccounts {
          id
          metadata {
            name
            annotations
          }
          updateTime
          displayName
        }
      }
    `,
    {
      transformer: (data: ConsoleListAccountsQuery) =>
        data.accounts_listAccounts,
      vars(_: ConsoleListAccountsQueryVariables) {},
    }
  ),

  getAccount: executor(
    gql`
      query Accounts_getAccount($accountName: String!) {
        accounts_getAccount(accountName: $accountName) {
          metadata {
            name
            annotations
          }
          updateTime
          contactEmail
          displayName
        }
      }
    `,
    {
      transformer: (data: ConsoleGetAccountQuery) => data.accounts_getAccount,
      vars(_: ConsoleGetAccountQueryVariables) {},
    }
  ),
});
