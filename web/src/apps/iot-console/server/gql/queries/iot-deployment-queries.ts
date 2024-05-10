import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  IotconsoleCreateIotDeploymentMutation,
  IotconsoleCreateIotDeploymentMutationVariables,
  IotconsoleGetIotDeploymentQuery,
  IotconsoleGetIotDeploymentQueryVariables,
  IotconsoleListIotDeploymentsQuery,
  IotconsoleListIotDeploymentsQueryVariables,
  IotconsoleUpdateIotDeploymentMutation,
  IotconsoleUpdateIotDeploymentMutationVariables,
  IotconsoleDeleteIotDeploymentMutation,
  IotconsoleDeleteIotDeploymentMutationVariables,
} from '~/root/src/generated/gql/server';

export type IDeployments = NN<
  IotconsoleListIotDeploymentsQuery['iot_listDeployments']
>;
export type IDeployment = NN<
  IotconsoleGetIotDeploymentQuery['iot_getDeployment']
>;

export const iotDeploymentQueries = (executor: IExecutor) => ({
  deleteIotDeployment: executor(
    gql`
      mutation Iot_deleteDeployment($projectName: String!, $name: String!) {
        iot_deleteDeployment(projectName: $projectName, name: $name)
      }
    `,
    {
      transformer: (data: IotconsoleDeleteIotDeploymentMutation) =>
        data.iot_deleteDeployment,
      vars(_: IotconsoleDeleteIotDeploymentMutationVariables) {},
    }
  ),
  createIotDeployment: executor(
    gql`
      mutation Iot_createDeployment(
        $projectName: String!
        $deployment: IOTDeploymentIn!
      ) {
        iot_createDeployment(
          projectName: $projectName
          deployment: $deployment
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: IotconsoleCreateIotDeploymentMutation) =>
        data.iot_createDeployment,
      vars(_: IotconsoleCreateIotDeploymentMutationVariables) {},
    }
  ),
  updateIotDeployment: executor(
    gql`
      mutation Iot_updateDeployment(
        $projectName: String!
        $deployment: IOTDeploymentIn!
      ) {
        iot_updateDeployment(
          projectName: $projectName
          deployment: $deployment
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: IotconsoleUpdateIotDeploymentMutation) =>
        data.iot_updateDeployment,
      vars(_: IotconsoleUpdateIotDeploymentMutationVariables) {},
    }
  ),
  getIotDeployment: executor(
    gql`
      query Iot_getDeployment($projectName: String!, $name: String!) {
        iot_getDeployment(projectName: $projectName, name: $name) {
          accountName
          CIDR
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
          exposedServices {
            ip
            name
          }
          id
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          # exposedDomains
          # exposedIps
          markedForDeletion
          name
          recordVersion
          updateTime
        }
      }
    `,
    {
      transformer: (data: IotconsoleGetIotDeploymentQuery) =>
        data.iot_getDeployment,
      vars(_: IotconsoleGetIotDeploymentQueryVariables) {},
    }
  ),
  listIotDeployments: executor(
    gql`
      query Iot_listDeployments(
        $search: SearchIOTDeployments
        $pq: CursorPaginationIn
        $projectName: String!
      ) {
        iot_listDeployments(
          projectName: $projectName
          search: $search
          pq: $pq
        ) {
          edges {
            cursor
            node {
              accountName
              CIDR
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              exposedServices {
                ip
                name
              }
              id
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              # exposedDomains
              # exposedIps
              markedForDeletion
              name
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
      transformer: (data: IotconsoleListIotDeploymentsQuery) => {
        return data.iot_listDeployments;
      },
      vars(_: IotconsoleListIotDeploymentsQueryVariables) {},
    }
  ),
});
