import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateAccountMutation,
  ConsoleCreateAccountMutationVariables,
  ConsoleDeleteAccountMutation,
  ConsoleDeleteAccountMutationVariables,
  ConsoleGetAccountQuery,
  ConsoleGetAccountQueryVariables,
  ConsoleGetAvailableKloudliteRegionsQuery,
  ConsoleGetAvailableKloudliteRegionsQueryVariables,
  ConsoleListAccountsQuery,
  ConsoleListAccountsQueryVariables,
  ConsoleUpdateAccountMutation,
  ConsoleUpdateAccountMutationVariables,
} from '~/root/src/generated/gql/server';

export type IAccounts = NN<ConsoleListAccountsQuery['accounts_listAccounts']>;
export type IAccount = NN<ConsoleGetAccountQuery['accounts_getAccount']>;

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
  getAvailableKloudliteRegions: executor(
    gql`
      query Accounts_availableKloudliteRegions {
        accounts_availableKloudliteRegions {
          displayName
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleGetAvailableKloudliteRegionsQuery) =>
        data.accounts_availableKloudliteRegions,
      vars(_: ConsoleGetAvailableKloudliteRegionsQueryVariables) {},
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
          kloudliteGatewayRegion
        }
      }
    `,
    {
      transformer: (data: ConsoleListAccountsQuery) =>
        data.accounts_listAccounts,
      vars(_: ConsoleListAccountsQueryVariables) {},
    }
  ),
  updateAccount: executor(
    gql`
      mutation Accounts_updateAccount($account: AccountIn!) {
        accounts_updateAccount(account: $account) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateAccountMutation) =>
        data.accounts_updateAccount,
      vars(_: ConsoleUpdateAccountMutationVariables) {},
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
          targetNamespace
          updateTime
          contactEmail
          displayName
          kloudliteGatewayRegion
        }
      }
    `,
    {
      transformer: (data: ConsoleGetAccountQuery) => data.accounts_getAccount,
      vars(_: ConsoleGetAccountQueryVariables) {},
    }
  ),
  deleteAccount: executor(
    gql`
      mutation Accounts_deleteAccount($accountName: String!) {
        accounts_deleteAccount(accountName: $accountName)
      }
    `,
    {
      transformer: (data: ConsoleDeleteAccountMutation) =>
        data.accounts_deleteAccount,
      vars(_: ConsoleDeleteAccountMutationVariables) {},
    }
  ),
});
