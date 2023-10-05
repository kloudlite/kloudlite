import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleListGithubInstalltionsQuery,
  ConsoleListGithubInstalltionsQueryVariables,
  ConsoleListGithubReposQuery,
  ConsoleListGithubReposQueryVariables,
} from '~/root/src/generated/gql/server';

export type IGithubRepos = NN<
  ConsoleListGithubReposQuery['cr_listGithubRepos']
>;
export type IGithubInstallations = NN<
  ConsoleListGithubInstalltionsQuery['cr_listGithubInstallations']
>;
// export type IProject = NN<Console['core_getProject']>;

export const gitQueries = (executor: IExecutor) => ({
  listGithubRepos: executor(
    gql`
      query Cr_listGithubRepos(
        $installationId: Int!
        $pagination: PaginationIn
      ) {
        cr_listGithubRepos(
          installationId: $installationId
          pagination: $pagination
        ) {
          repositories {
            archived
            clone_url
            created_at
            default_branch
            description
            disabled
            full_name
            git_url
            gitignore_template
            html_url
            id
            language
            master_branch
            mirror_url
            name
            node_id
            permissions
            private
            pushed_at
            size
            team_id
            updated_at
            visibility
          }
          total_count
        }
      }
    `,
    {
      transformer: (data: ConsoleListGithubReposQuery) =>
        data.cr_listGithubRepos,
      vars(_: ConsoleListGithubReposQueryVariables) {},
    }
  ),
  listGithubInstalltions: executor(
    gql`
      query Cr_listGithubInstallations($pagination: PaginationIn) {
        cr_listGithubInstallations(pagination: $pagination) {
          account {
            avatar_url
            id
            login
            node_id
            type
          }
          app_id
          id
          node_id
          repositories_url
          target_id
          target_type
        }
      }
    `,
    {
      transformer: (data: ConsoleListGithubInstalltionsQuery) =>
        data.cr_listGithubInstallations,
      vars(_: ConsoleListGithubInstalltionsQueryVariables) {},
    }
  ),
});
