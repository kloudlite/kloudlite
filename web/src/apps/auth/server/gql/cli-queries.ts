/* eslint-disable camelcase */
import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  AuthCli_CreateRemoteLoginMutationVariables,
  AuthCli_CreateRemoteLoginMutation,
  AuthCli_GetRemoteLoginQueryVariables,
  AuthCli_GetRemoteLoginQuery,
  AuthCli_GetCurrentUserQuery,
  AuthCli_GetCurrentUserQueryVariables,
  AuthCli_ListAccountsQuery,
  AuthCli_ListAccountsQueryVariables,
  AuthCli_ListClustersQuery,
  AuthCli_ListClustersQueryVariables,
  AuthCli_GetKubeConfigQuery,
  AuthCli_GetKubeConfigQueryVariables,
} from '~/root/src/generated/gql/server';

export const cliQueries = (executor: IExecutor) => ({
  cli_listEnvironments: executor(
    gql`
      query Core_listProjects($project: ProjectId!, $pq: CursorPaginationIn) {
        core_listEnvironments(project: $project, pq: $pq) {
          edges {
            node {
              displayName
              markedForDeletion
              metadata {
                name
                namespace
              }
              spec {
                isEnvironment
                projectName
                targetNamespace
              }
              status {
                isReady
                message {
                  RawMessage
                }
              }
            }
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.core_listEnvironments,
      vars: (_: any) => {},
    }
  ),

  cli_listProjects: executor(
    gql`
      query Core_listProjects($clusterName: String, $pq: CursorPaginationIn) {
        core_listProjects(clusterName: $clusterName, pq: $pq) {
          edges {
            node {
              displayName
              markedForDeletion
              metadata {
                name
                namespace
              }
              status {
                isReady
                message {
                  RawMessage
                }
              }
            }
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.core_listProjects,
      vars: (_: any) => {},
    }
  ),

  cli_getKubeConfig: executor(
    gql`
      query Infra_getCluster($name: String!) {
        infra_getCluster(name: $name) {
          adminKubeconfig {
            encoding
            value
          }
          status {
            isReady
          }
        }
      }
    `,
    {
      transformer: (data: AuthCli_GetKubeConfigQuery) => data.infra_getCluster,
      vars(_: AuthCli_GetKubeConfigQueryVariables) {},
    }
  ),
  cli_listClusters: executor(
    gql`
      query Node($pagination: CursorPaginationIn) {
        infra_listClusters(pagination: $pagination) {
          edges {
            node {
              displayName
              metadata {
                name
              }
              status {
                isReady
              }
            }
          }
        }
      }
    `,
    {
      transformer(data: AuthCli_ListClustersQuery) {
        return data.infra_listClusters;
      },
      vars(_: AuthCli_ListClustersQueryVariables) {},
    }
  ),
  cli_listAccounts: executor(
    gql`
      query Accounts_listAccounts {
        accounts_listAccounts {
          metadata {
            name
          }
          displayName
        }
      }
    `,
    {
      transformer(data: AuthCli_ListAccountsQuery) {
        return data.accounts_listAccounts;
      },
      vars(_: AuthCli_ListAccountsQueryVariables) {},
    }
  ),
  cli_getCurrentUser: executor(
    gql`
      query Auth_me {
        auth_me {
          id
          email
          name
        }
      }
    `,
    {
      transformer(data: AuthCli_GetCurrentUserQuery) {
        return data.auth_me;
      },
      vars(_: AuthCli_GetCurrentUserQueryVariables) {},
    }
  ),

  cli_createRemoteLogin: executor(
    gql`
      mutation Auth_createRemoteLogin($secret: String) {
        auth_createRemoteLogin(secret: $secret)
      }
    `,
    {
      transformer: (data: AuthCli_CreateRemoteLoginMutation) =>
        data.auth_createRemoteLogin,
      vars(_: AuthCli_CreateRemoteLoginMutationVariables) {},
    }
  ),

  cli_getRemoteLogin: executor(
    gql`
      query Auth_getRemoteLogin($loginId: String!, $secret: String!) {
        auth_getRemoteLogin(loginId: $loginId, secret: $secret) {
          authHeader
          status
        }
      }
    `,
    {
      transformer: (data: AuthCli_GetRemoteLoginQuery) =>
        data.auth_getRemoteLogin,
      vars(_: AuthCli_GetRemoteLoginQueryVariables) {},
    }
  ),
});
