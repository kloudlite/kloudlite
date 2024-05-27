import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleDeleteAccountInvitationMutation,
  ConsoleDeleteAccountInvitationMutationVariables,
  ConsoleInviteMembersForAccountMutation,
  ConsoleInviteMembersForAccountMutationVariables,
  ConsoleListInvitationsForAccountQuery,
  ConsoleListInvitationsForAccountQueryVariables,
  ConsoleListMembershipsForAccountQuery,
  ConsoleListMembershipsForAccountQueryVariables,
  ConsoleUpdateAccountMembershipMutation,
  ConsoleUpdateAccountMembershipMutationVariables,
  ConsoleListInvitationsForUserQuery,
  ConsoleListInvitationsForUserQueryVariables,
  ConsoleAcceptInvitationMutation,
  ConsoleAcceptInvitationMutationVariables,
  ConsoleRejectInvitationMutation,
  ConsoleRejectInvitationMutationVariables,
  ConsoleDeleteAccountMembershipMutation,
  ConsoleDeleteAccountMembershipMutationVariables,
  ConsoleVerifyInviteCodeMutation,
  ConsoleVerifyInviteCodeMutationVariables,
} from '~/root/src/generated/gql/server';

export type IInvites = NN<
  ConsoleListInvitationsForUserQuery['accounts_listInvitationsForUser']
>;

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
        $invitations: [InvitationIn!]!
      ) {
        accounts_inviteMembers(
          accountName: $accountName
          invitations: $invitations
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
  listInvitationsForUser: executor(
    gql`
      query Accounts_listInvitationsForUser($onlyPending: Boolean!) {
        accounts_listInvitationsForUser(onlyPending: $onlyPending) {
          accountName
          id
          updateTime
          inviteToken
        }
      }
    `,
    {
      transformer(data: ConsoleListInvitationsForUserQuery) {
        return data.accounts_listInvitationsForUser;
      },
      vars(_: ConsoleListInvitationsForUserQueryVariables) {},
    }
  ),
  acceptInvitation: executor(
    gql`
      mutation Accounts_acceptInvitation(
        $accountName: String!
        $inviteToken: String!
      ) {
        accounts_acceptInvitation(
          accountName: $accountName
          inviteToken: $inviteToken
        )
      }
    `,
    {
      transformer(data: ConsoleAcceptInvitationMutation) {
        return data.accounts_acceptInvitation;
      },
      vars(_: ConsoleAcceptInvitationMutationVariables) {},
    }
  ),
  rejectInvitation: executor(
    gql`
      mutation Accounts_rejectInvitation(
        $accountName: String!
        $inviteToken: String!
      ) {
        accounts_rejectInvitation(
          accountName: $accountName
          inviteToken: $inviteToken
        )
      }
    `,
    {
      transformer(data: ConsoleRejectInvitationMutation) {
        return data.accounts_rejectInvitation;
      },
      vars(_: ConsoleRejectInvitationMutationVariables) {},
    }
  ),
  updateAccountMembership: executor(
    gql`
      mutation Accounts_updateAccountMembership(
        $accountName: String!
        $memberId: ID!
        $role: Github__com___kloudlite___api___apps___iam___types__Role!
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
  deleteAccountMembership: executor(
    gql`
      mutation Accounts_removeAccountMembership(
        $accountName: String!
        $memberId: ID!
      ) {
        accounts_removeAccountMembership(
          accountName: $accountName
          memberId: $memberId
        )
      }
    `,
    {
      transformer(data: ConsoleDeleteAccountMembershipMutation) {
        return data.accounts_removeAccountMembership;
      },
      vars(_: ConsoleDeleteAccountMembershipMutationVariables) {},
    }
  ),

  verifyInviteCode: executor(
    gql`
      mutation Auth_verifyInviteCode($invitationCode: String!) {
        auth_verifyInviteCode(invitationCode: $invitationCode)
      }
    `,
    {
      transformer: (data: ConsoleVerifyInviteCodeMutation) =>
        data.auth_verifyInviteCode,
      vars(_: ConsoleVerifyInviteCodeMutationVariables) {},
    }
  ),
});
