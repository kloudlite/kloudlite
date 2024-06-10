import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleUpdateNotificationConfigMutation,
  ConsoleUpdateNotificationConfigMutationVariables,
  ConsoleUpdateSubscriptionConfigMutation,
  ConsoleUpdateSubscriptionConfigMutationVariables,
  ConsoleMarkAllNotificationAsReadMutation,
  ConsoleMarkAllNotificationAsReadMutationVariables,
  ConsoleGetNotificationConfigQuery,
  ConsoleGetNotificationConfigQueryVariables,
  ConsoleGetSubscriptionConfigQuery,
  ConsoleGetSubscriptionConfigQueryVariables,
  ConsoleListNotificationsQuery,
  ConsoleListNotificationsQueryVariables,
} from '~/root/src/generated/gql/server';

export type ICommsNotifications = NN<
  ConsoleListNotificationsQuery['comms_listNotifications']
>;
// export type ICommsNotification = NN<
//   ConsoleGetByokClusterQuery['infra_getBYOKCluster']
// >;

export const commsNotificationQueries = (executor: IExecutor) => ({
  updateNotificationConfig: executor(
    gql`
      mutation Comms_updateNotificationConfig($config: NotificationConfIn!) {
        comms_updateNotificationConfig(config: $config) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateNotificationConfigMutation) =>
        data.comms_updateNotificationConfig,
      vars(_: ConsoleUpdateNotificationConfigMutationVariables) {},
    }
  ),
  updateSubscriptionConfig: executor(
    gql`
      mutation Comms_updateSubscriptionConfig(
        $config: SubscriptionIn!
        $commsUpdateSubscriptionConfigId: ID!
      ) {
        comms_updateSubscriptionConfig(
          config: $config
          id: $commsUpdateSubscriptionConfigId
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateSubscriptionConfigMutation) =>
        data.comms_updateSubscriptionConfig,
      vars(_: ConsoleUpdateSubscriptionConfigMutationVariables) {},
    }
  ),
  markAllNotificationAsRead: executor(
    gql`
      mutation Mutation {
        comms_markAllNotificationAsRead
      }
    `,
    {
      transformer: (data: ConsoleMarkAllNotificationAsReadMutation) =>
        data.comms_markAllNotificationAsRead,
      vars(_: ConsoleMarkAllNotificationAsReadMutationVariables) {},
    }
  ),
  getNotificationConfig: executor(
    gql`
      query Comms_getNotificationConfig {
        comms_getNotificationConfig {
          accountName
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          email {
            enabled
            mailAddress
          }
          id
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          markedForDeletion
          recordVersion
          slack {
            enabled
            url
          }
          telegram {
            chatId
            enabled
            token
          }
          updateTime
          webhook {
            enabled
            url
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleGetNotificationConfigQuery) =>
        data.comms_getNotificationConfig,
      vars(_: ConsoleGetNotificationConfigQueryVariables) {},
    }
  ),
  getSubscriptionConfig: executor(
    gql`
      query Comms_getSubscriptionConfig($commsGetSubscriptionConfigId: ID!) {
        comms_getSubscriptionConfig(id: $commsGetSubscriptionConfigId) {
          accountName
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          enabled
          id
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          mailAddress
          markedForDeletion
          recordVersion
          updateTime
        }
      }
    `,
    {
      transformer: (data: ConsoleGetSubscriptionConfigQuery) =>
        data.comms_getSubscriptionConfig,
      vars(_: ConsoleGetSubscriptionConfigQueryVariables) {},
    }
  ),
  listNotifications: executor(
    gql`
      query Comms_listNotifications($pagination: CursorPaginationIn) {
        comms_listNotifications(pagination: $pagination) {
          edges {
            cursor
            node {
              accountName
              content {
                body
                image
                link
                subject
                title
              }
              creationTime
              id
              markedForDeletion
              notificationType
              priority
              read
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
      transformer: (data: ConsoleListNotificationsQuery) => {
        return data.comms_listNotifications;
      },
      vars(_: ConsoleListNotificationsQueryVariables) {},
    }
  ),
});
