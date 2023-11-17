import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleGetLoginsQuery,
  ConsoleGetLoginsQueryVariables,
  ConsoleListGithubBranchesQuery,
  ConsoleListGithubBranchesQueryVariables,
  ConsoleListGithubInstalltionsQuery,
  ConsoleListGithubInstalltionsQueryVariables,
  ConsoleListGithubReposQuery,
  ConsoleListGithubReposQueryVariables,
  ConsoleListGitlabBranchesQuery,
  ConsoleListGitlabBranchesQueryVariables,
  ConsoleListGitlabGroupsQuery,
  ConsoleListGitlabGroupsQueryVariables,
  ConsoleListGitlabReposQuery,
  ConsoleListGitlabReposQueryVariables,
  ConsoleLoginUrlsQuery,
  ConsoleLoginUrlsQueryVariables,
  ConsoleSearchGithubReposQuery,
  ConsoleSearchGithubReposQueryVariables,
} from '~/root/src/generated/gql/server';

export type IGithubRepos = NN<
  ConsoleListGithubReposQuery['cr_listGithubRepos']
>;
export type IGithubInstallations = NN<
  ConsoleListGithubInstalltionsQuery['cr_listGithubInstallations']
>;
export type IGitlabGroups = NN<
  ConsoleListGitlabGroupsQuery['cr_listGitlabGroups']
>;
// export type IProject = NN<Console['core_getProject']>;

export const gitQueries = (executor: IExecutor) => ({
  getLogins: executor(
    gql`
      query Auth_me {
        auth_me {
          providerGithub
          providerGitlab
        }
      }
    `,
    {
      transformer(data: ConsoleGetLoginsQuery) {
        return data.auth_me;
      },
      vars(_: ConsoleGetLoginsQueryVariables) {},
    }
  ),

  loginUrls: executor(
    gql`
      query Query {
        githubLoginUrl: oAuth_requestLogin(
          provider: "github"
          state: "redirect:add-provider"
        )
        gitlabLoginUrl: oAuth_requestLogin(
          provider: "gitlab"
          state: "redirect:add-provider"
        )
      }
    `,
    {
      transformer: (data: ConsoleLoginUrlsQuery) => data,
      vars(_: ConsoleLoginUrlsQueryVariables) {},
    }
  ),

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
            cloneUrl
            defaultBranch
            fullName
            private
            updatedAt
          }
          totalCount
        }
      }
    `,
    {
      transformer: (data: ConsoleListGithubReposQuery) => {
        return data.cr_listGithubRepos?.repositories.map((r) => {
          return {
            name: r.fullName || '',
            updatedAt: r.updatedAt,
            private: r.private || true,
            url: r.cloneUrl || '',
          };
        });
      },
      vars(_: ConsoleListGithubReposQueryVariables) {},
    }
  ),
  listGithubInstalltions: executor(
    gql`
      query Cr_listGithubInstallations($pagination: PaginationIn) {
        cr_listGithubInstallations(pagination: $pagination) {
          account {
            avatarUrl
            id
            login
            nodeId
            type
          }
          appId
          id
          nodeId
          repositoriesUrl
          targetId
          targetType
        }
      }
    `,
    {
      transformer: (data: ConsoleListGithubInstalltionsQuery) =>
        data.cr_listGithubInstallations,
      vars(_: ConsoleListGithubInstalltionsQueryVariables) {},
    }
  ),
  listGithubBranches: executor(
    gql`
      query Cr_listGithubBranches(
        $repoUrl: String!
        $pagination: PaginationIn
      ) {
        cr_listGithubBranches(repoUrl: $repoUrl, pagination: $pagination) {
          name
        }
      }
    `,
    {
      transformer: (data: ConsoleListGithubBranchesQuery) =>
        data.cr_listGithubBranches,
      vars(_: ConsoleListGithubBranchesQueryVariables) {},
    }
  ),
  searchGithubRepos: executor(
    gql`
      query Cr_searchGithubRepos(
        $organization: String!
        $search: String!
        $pagination: PaginationIn
      ) {
        cr_searchGithubRepos(
          organization: $organization
          search: $search
          pagination: $pagination
        ) {
          repositories {
            cloneUrl
            defaultBranch
            fullName
            private
            updatedAt
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleSearchGithubReposQuery) => {
        return data.cr_searchGithubRepos?.repositories.map((r) => {
          return {
            name: r.fullName || '',
            updatedAt: r.updatedAt,
            private: r.private || true,
            url: r.cloneUrl || '',
          };
        });
      },
      vars(_: ConsoleSearchGithubReposQueryVariables) {},
    }
  ),
  listGitlabGroups: executor(
    gql`
      query Cr_listGitlabGroups($query: String, $pagination: PaginationIn) {
        cr_listGitlabGroups(query: $query, pagination: $pagination) {
          fullName
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleListGitlabGroupsQuery) =>
        data.cr_listGitlabGroups,
      vars(_: ConsoleListGitlabGroupsQueryVariables) {},
    }
  ),
  listGitlabRepos: executor(
    gql`
      query Cr_listGitlabRepositories(
        $query: String
        $pagination: PaginationIn
        $groupId: String!
      ) {
        cr_listGitlabRepositories(
          query: $query
          pagination: $pagination
          groupId: $groupId
        ) {
          createdAt
          name
          id
          public
          httpUrlToRepo
        }
      }
    `,
    {
      transformer: (data: ConsoleListGitlabReposQuery) => {
        return data.cr_listGitlabRepositories?.map((r) => {
          return {
            name: r.name || '',
            updatedAt: r.createdAt || '',
            private: r.public || true,
            url: `${r.id}` || '',
          };
        });
      },
      vars(_: ConsoleListGitlabReposQueryVariables) {},
    }
  ),
  listGitlabBranches: executor(
    gql`
      query Cr_listGitlabBranches(
        $repoId: String!
        $query: String
        $pagination: PaginationIn
      ) {
        cr_listGitlabBranches(
          repoId: $repoId
          query: $query
          pagination: $pagination
        ) {
          name
          protected
        }
      }
    `,
    {
      transformer: (data: ConsoleListGitlabBranchesQuery) =>
        data.cr_listGitlabBranches,
      vars(_: ConsoleListGitlabBranchesQueryVariables) {},
    }
  ),
});
