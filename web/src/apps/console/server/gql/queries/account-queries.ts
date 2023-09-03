import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';

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
      transformer(data) {},
      vars(variables) {},
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
      transformer(data) {},
      vars(variables) {},
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
      transformer(data) {},
      vars(variables) {},
    }
  ),

  // inviteUser: executor(gql`
  //   mutation Finance_inviteUser(
  //     $accountName: String!
  //     $email: String!
  //     $role: String!
  //   ) {
  //     finance_inviteUser(accountName: $accountName, email: $email, role: $role)
  //   }
  // `),

  // listUsers: executor(gql`
  //   query Finance_listInvitations($accountName: String!) {
  //     finance_listInvitations(accountName: $accountName) {
  //       accepted
  //       user {
  //         id
  //         name
  //         verified
  //         email
  //         avatar
  //       }
  //       role
  //     }
  //   }
  // `),
});
