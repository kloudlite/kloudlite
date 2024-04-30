import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  IotconsoleCreateIotDeviceBlueprintMutation,
  IotconsoleCreateIotDeviceBlueprintMutationVariables,
  IotconsoleGetIotDeviceBlueprintQuery,
  IotconsoleGetIotDeviceBlueprintQueryVariables,
  IotconsoleListIotDeviceBlueprintsQuery,
  IotconsoleListIotDeviceBlueprintsQueryVariables,
  IotconsoleUpdateIotDeviceBlueprintMutation,
  IotconsoleUpdateIotDeviceBlueprintMutationVariables,
  IotconsoleDeleteIotDeviceBlueprintMutation,
  IotconsoleDeleteIotDeviceBlueprintMutationVariables,
} from '~/root/src/generated/gql/server';

export type IDeviceBlueprints = NN<
  IotconsoleListIotDeviceBlueprintsQuery['iot_listDeviceBlueprints']
>;
export type IDeviceBlueprint = NN<
  IotconsoleGetIotDeviceBlueprintQuery['iot_getDeviceBlueprint']
>;

export const iotDeviceBlueprintQueries = (executor: IExecutor) => ({
  deleteIotDeviceBlueprint: executor(
    gql`
      mutation Iot_deleteDeviceBlueprint(
        $projectName: String!
        $name: String!
      ) {
        iot_deleteDeviceBlueprint(projectName: $projectName, name: $name)
      }
    `,
    {
      transformer: (data: IotconsoleDeleteIotDeviceBlueprintMutation) =>
        data.iot_deleteDeviceBlueprint,
      vars(_: IotconsoleDeleteIotDeviceBlueprintMutationVariables) {},
    }
  ),
  createIotDeviceBlueprint: executor(
    gql`
      mutation Iot_createDeviceBlueprint(
        $deviceBlueprint: IOTDeviceBlueprintIn!
        $projectName: String!
      ) {
        iot_createDeviceBlueprint(
          projectName: $projectName
          deviceBlueprint: $deviceBlueprint
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: IotconsoleCreateIotDeviceBlueprintMutation) =>
        data.iot_createDeviceBlueprint,
      vars(_: IotconsoleCreateIotDeviceBlueprintMutationVariables) {},
    }
  ),
  updateIotDeviceBlueprint: executor(
    gql`
      mutation Iot_updateDeviceBlueprint(
        $deviceBlueprint: IOTDeviceBlueprintIn!
        $projectName: String!
      ) {
        iot_updateDeviceBlueprint(
          projectName: $projectName
          deviceBlueprint: $deviceBlueprint
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: IotconsoleUpdateIotDeviceBlueprintMutation) =>
        data.iot_updateDeviceBlueprint,
      vars(_: IotconsoleUpdateIotDeviceBlueprintMutationVariables) {},
    }
  ),
  getIotDeviceBlueprint: executor(
    gql`
      query Iot_getDeviceBlueprint($projectName: String!, $name: String!) {
        iot_getDeviceBlueprint(projectName: $projectName, name: $name) {
          accountName
          bluePrintType
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
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
          version
        }
      }
    `,
    {
      transformer: (data: IotconsoleGetIotDeviceBlueprintQuery) =>
        data.iot_getDeviceBlueprint,
      vars(_: IotconsoleGetIotDeviceBlueprintQueryVariables) {},
    }
  ),
  listIotDeviceBlueprints: executor(
    gql`
      query Iot_listDeviceBlueprints(
        $search: SearchIOTDeviceBlueprints
        $pq: CursorPaginationIn
        $projectName: String!
      ) {
        iot_listDeviceBlueprints(
          projectName: $projectName
          search: $search
          pq: $pq
        ) {
          edges {
            cursor
            node {
              accountName
              bluePrintType
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
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
              version
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
      transformer: (data: IotconsoleListIotDeviceBlueprintsQuery) => {
        return data.iot_listDeviceBlueprints;
      },
      vars(_: IotconsoleListIotDeviceBlueprintsQueryVariables) {},
    }
  ),
});
