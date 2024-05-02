import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleListGlobalVpnDevicesQuery,
  ConsoleListGlobalVpnDevicesQueryVariables,
} from '~/root/src/generated/gql/server';

export type IGlobalVpnDevices = NN<
  ConsoleListGlobalVpnDevicesQuery['infra_listGlobalVPNDevices']
>;

export const globalVpnQueries = (executor: IExecutor) => ({
  listGlobalVpnDevices: executor(
    gql`
      query Infra_listGlobalVPNDevices(
        $gvpn: String!
        $search: SearchGlobalVPNDevices
        $pagination: CursorPaginationIn
      ) {
        infra_listGlobalVPNDevices(
          gvpn: $gvpn
          search: $search
          pagination: $pagination
        ) {
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
              displayName
              globalVPNName
              id
              ipAddr
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              markedForDeletion
              metadata {
                annotations
                creationTimestamp
                deletionTimestamp
                generation
                labels
                name
                namespace
              }
              privateKey
              publicKey
              publiEndpoint
              recordVersion
              updateTime
              wireguardConfig {
                value
                encoding
              }
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
      transformer: (data: ConsoleListGlobalVpnDevicesQuery) =>
        data.infra_listGlobalVPNDevices,
      vars(_: ConsoleListGlobalVpnDevicesQueryVariables) {},
    }
  ),
});
