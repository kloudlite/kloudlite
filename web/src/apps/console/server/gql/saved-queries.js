import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const GQLServerHandler = ({ headers }) => {
  const executor = ExecuteQueryWithContext(headers);
  return {
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

    createAccount: executor(
      gql`
        mutation Finance_activateAccount(
          $name: String!
          $displayName: String!
        ) {
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

    whoAmI: executor(
      gql`
        query Auth_me {
          auth_me {
            id
            email
          }
        }
      `,
      {
        dataPath: 'auth_me',
        transformer: (me) => ({ me }),
      }
    ),
  };
};
