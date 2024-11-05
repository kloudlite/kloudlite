import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  IotconsoleCreateIotProjectMutation,
  IotconsoleCreateIotProjectMutationVariables,
  IotconsoleGetIotProjectQuery,
  IotconsoleGetIotProjectQueryVariables,
  IotconsoleListIotProjectsQuery,
  IotconsoleListIotProjectsQueryVariables,
  IotconsoleUpdateIotProjectMutation,
  IotconsoleUpdateIotProjectMutationVariables,
  IotconsoleDeleteIotProjectMutation,
  IotconsoleDeleteIotProjectMutationVariables,
} from '~/root/src/generated/gql/server';

export type IProjects = NN<IotconsoleListIotProjectsQuery['iot_listProjects']>;
export type IProject = NN<IotconsoleGetIotProjectQuery['iot_getProject']>;

export const iotProjectQueries = (executor: IExecutor) => ({
  deleteIotProject: executor(
    gql`
      mutation Iot_deleteProject($name: String!) {
        iot_deleteProject(name: $name)
      }
    `,
    {
      transformer: (data: IotconsoleDeleteIotProjectMutation) =>
        data.iot_deleteProject,
      vars(_: IotconsoleDeleteIotProjectMutationVariables) {},
    }
  ),
  createIotProject: executor(
    gql`
      mutation Iot_createProject($project: IOTProjectIn!) {
        iot_createProject(project: $project) {
          id
        }
      }
    `,
    {
      transformer: (data: IotconsoleCreateIotProjectMutation) =>
        data.iot_createProject,
      vars(_: IotconsoleCreateIotProjectMutationVariables) {},
    }
  ),
  updateIotProject: executor(
    gql`
      mutation Iot_updateProject($project: IOTProjectIn!) {
        iot_updateProject(project: $project) {
          id
        }
      }
    `,
    {
      transformer: (data: IotconsoleUpdateIotProjectMutation) =>
        data.iot_updateProject,
      vars(_: IotconsoleUpdateIotProjectMutationVariables) {},
    }
  ),
  getIotProject: executor(
    gql`
      query Iot_getProject($name: String!) {
        iot_getProject(name: $name) {
          accountName
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
        }
      }
    `,
    {
      transformer: (data: IotconsoleGetIotProjectQuery) => data.iot_getProject,
      vars(_: IotconsoleGetIotProjectQueryVariables) {},
    }
  ),
  listIotProjects: executor(
    gql`
      query Iot_listProjects(
        $search: SearchIOTProjects
        $pq: CursorPaginationIn
      ) {
        iot_listProjects(search: $search, pq: $pq) {
          edges {
            node {
              displayName
              name
              creationTime
              markedForDeletion
              updateTime
              createdBy {
                userEmail
                userName
                userId
              }
              lastUpdatedBy {
                userEmail
                userName
                userId
              }
            }
            cursor
          }
          pageInfo {
            endCursor
            hasNextPage
            hasPrevPage
            startCursor
          }
          totalCount
        }
      }
    `,
    {
      transformer: (data: IotconsoleListIotProjectsQuery) => {
        return data.iot_listProjects;
      },
      vars(_: IotconsoleListIotProjectsQueryVariables) {},
    }
  ),
});
