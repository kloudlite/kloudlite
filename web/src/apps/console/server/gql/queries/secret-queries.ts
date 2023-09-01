import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGqlReturn } from '~/root/lib/types/common';

export interface IGQLMethodsSecret {
  listSecrets: (variables?: any) => IGqlReturn<any>;
  createSecret: (variables?: any) => IGqlReturn<any>;
  getSecret: (variables?: any) => IGqlReturn<any>;
}

export const secretQueries = (executor: IExecutor): IGQLMethodsSecret => ({
  listSecrets: executor(
    gql`
      query Core_listSecrets(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $pq: CursorPaginationIn
        $search: SearchSecrets
      ) {
        core_listSecrets(
          project: $project
          scope: $scope
          pq: $pq
          search: $search
        ) {
          pageInfo {
            startCursor
            hasPreviousPage
            hasNextPage
            endCursor
          }
          totalCount
          edges {
            node {
              metadata {
                namespace
                name
                annotations
                labels
              }
              updateTime
              stringData
            }
          }
        }
      }
    `,
    {
      dataPath: 'core_listSecrets',
    }
  ),
  createSecret: executor(gql`
    mutation Mutation($secret: SecretIn!) {
      core_createSecret(secret: $secret) {
        id
      }
    }
  `),

  getSecret: executor(
    gql`
      query Core_getSecret(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $name: String!
      ) {
        core_getSecret(project: $project, scope: $scope, name: $name) {
          stringData
          updateTime
          displayName
          metadata {
            name
            namespace
            annotations
            labels
          }
        }
      }
    `,
    {
      dataPath: 'core_getSecret',
    }
  ),
});
