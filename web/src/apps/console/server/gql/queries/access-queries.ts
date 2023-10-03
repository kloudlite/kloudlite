import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleDeleteAccountInvitationMutation,
  ConsoleDeleteAccountInvitationMutationVariables,
  ConsoleGetAccountQuery,
  ConsoleInviteMembersForAccountMutation,
  ConsoleInviteMembersForAccountMutationVariables,
  ConsoleListAccountsQuery,
  ConsoleListInvitationsForAccountQuery,
  ConsoleListInvitationsForAccountQueryVariables,
  ConsoleListMembershipsForAccountQuery,
  ConsoleListMembershipsForAccountQueryVariables,
  ConsoleUpdateAccountMembershipMutation,
  ConsoleUpdateAccountMembershipMutationVariables,
} from '~/root/src/generated/gql/server';

export type IAccounts = NN<ConsoleListAccountsQuery['accounts_listAccounts']>;
export type IAccount = NN<ConsoleGetAccountQuery['accounts_getAccount']>;

export const accessQueries = (executor: IExecutor) => ({
  listInvitationsForAccount: executor(
    gql`
      query Core_getVPNDevice($accountName: String!) {
        accounts_listInvitations(accountName: $accountName) {
          accepted
          accountName
          creationTime
          id
          inviteToken
          invitedBy
          markedForDeletion
          recordVersion
          rejected
          updateTime
          userEmail
          userName
          userRole
        }
      }
    `,
    {
      transformer(data: ConsoleListInvitationsForAccountQuery) {
        return data.accounts_listInvitations;
      },
      vars(_: ConsoleListInvitationsForAccountQueryVariables) {},
    }
  ),
  listMembershipsForAccount: executor(
    gql`
      query User($accountName: String!) {
        accounts_listMembershipsForAccount(accountName: $accountName) {
          user {
            verified
            name
            joined
            email
          }
          role
        }
      }
    `,
    {
      transformer: (data: ConsoleListMembershipsForAccountQuery) =>
        data.accounts_listMembershipsForAccount,
      vars(_: ConsoleListMembershipsForAccountQueryVariables) {},
    }
  ),

  deleteAccountInvitation: executor(
    gql`
      mutation Mutation($accountName: String!, $invitationId: String!) {
        accounts_deleteInvitation(
          accountName: $accountName
          invitationId: $invitationId
        )
      }
    `,
    {
      transformer(data: ConsoleDeleteAccountInvitationMutation) {
        return data.accounts_deleteInvitation;
      },
      vars(_: ConsoleDeleteAccountInvitationMutationVariables) {},
    }
  ),
  inviteMembersForAccount: executor(
    gql`
      mutation Accounts_inviteMembers(
        $accountName: String!
        $invitation: [InvitationIn!]!
      ) {
        accounts_inviteMembers(
          accountName: $accountName
          invitation: $invitation
        ) {
          id
        }
      }
    `,
    {
      transformer(data: ConsoleInviteMembersForAccountMutation) {
        return data.accounts_inviteMembers;
      },
      vars(_: ConsoleInviteMembersForAccountMutationVariables) {},
    }
  ),
  updateAccountMembership: executor(
    gql`
      mutation Mutation(
        $accountName: String!
        $memberId: ID!
        $role: Kloudlite_io__apps__iam__types_Role!
      ) {
        accounts_updateAccountMembership(
          accountName: $accountName
          memberId: $memberId
          role: $role
        )
      }
    `,
    {
      transformer(data: ConsoleUpdateAccountMembershipMutation) {
        return data.accounts_updateAccountMembership;
      },
      vars(_: ConsoleUpdateAccountMembershipMutationVariables) {},
    }
  ),
});
