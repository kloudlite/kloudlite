import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const environmentQueries = (executor = ExecuteQueryWithContext({})) => ({
  getEnvironment: executor(
    gql`
      query Core_getEnvironment($project: ProjectId!, $name: String!) {
        core_getEnvironment(project: $project, name: $name) {
          spec {
            targetNamespace
            projectName
          }
          updateTime
          metadata {
            namespace
            name
            annotations
            labels
          }
        }
      }
    `,
    {
      dataPath: 'core_getEnvironment',
    }
  ),
});
