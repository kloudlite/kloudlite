import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const accountQueries = (executor = ExecuteQueryWithContext({})) => ({
  createAccount: executor(
    gql`
      mutation Finance_activateAccount($name: String!, $displayName: String!) {
        finance_createAccount(name: $name, displayName: $displayName) {
          isActive
        }
      }
    `,
    {
      dataPath: 'finance_createAccount',
    }
  ),

  listAccounts: executor(
    gql`
      query Finance_listAccounts {
        finance_listAccounts {
          contactEmail
          isActive
          name
          displayName
          readableId
        }
      }
    `,
    {
      dataPath: 'finance_listAccounts',
    }
  ),

  getAccount: executor(
    gql`
      query Finance_account($accountName: String!) {
        finance_account(accountName: $accountName) {
          isActive
          readableId
          name
          displayName
          contactEmail
        }
      }
    `,
    {
      dataPath: 'finance_account',
    }
  ),

  inviteUser: executor(gql`
    mutation Finance_inviteUser(
      $accountName: String!
      $email: String!
      $role: String!
    ) {
      finance_inviteUser(accountName: $accountName, email: $email, role: $role)
    }
  `),

  listUsers: executor(gql`
    query Finance_listInvitations($accountName: String!) {
      finance_listInvitations(accountName: $accountName) {
        accepted
        user {
          id
          name
          verified
          email
          avatar
        }
        role
      }
    }
  `),
});
