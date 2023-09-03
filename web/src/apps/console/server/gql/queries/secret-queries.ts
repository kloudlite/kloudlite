import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';

export const secretQueries = (executor: IExecutor) => ({
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
      transformer(data) {},
      vars(variables) {},
    }
  ),
  createSecret: executor(
    gql`
      mutation createSecret($secret: SecretIn!) {
        core_createSecret(secret: $secret) {
          id
        }
      }
    `,
    {
      transformer(data) {},
      vars(variables) {},
    }
  ),

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
      transformer(data) {},
      vars(variables) {},
    }
  ),
});
