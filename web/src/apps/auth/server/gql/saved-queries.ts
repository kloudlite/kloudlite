import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGQLServerProps } from '~/root/lib/types/common';
import {
  AuthAddOauthCredientialsMutation,
  AuthLoginMutation,
  AuthLoginMutationVariables,
  AuthLoginPageInitUrlsQuery,
  AuthLoginPageInitUrlsQueryVariables,
  AuthLogoutMutation,
  AuthLogoutMutationVariables,
  AuthOauthLoginMutation,
  AuthOauthLoginMutationVariables,
  AuthRequestResetPasswordMutation,
  AuthRequestResetPasswordMutationVariables,
  AuthResetPasswordMutation,
  AuthResetPasswordMutationVariables,
  AuthSignUpWithEmailMutation,
  AuthSignUpWithEmailMutationVariables,
  AuthVerifyEmailMutation,
  AuthVerifyEmailMutationVariables,
  AuthWhoAmIQuery,
  AuthWhoAmIQueryVariables,
  AuthCheckOauthEnabledQuery,
  AuthCheckOauthEnabledQueryVariables,
} from '~/root/src/generated/gql/server';

export const GQLServerHandler = ({ headers, cookies }: IGQLServerProps) => {
  const executor = ExecuteQueryWithContext(headers, cookies);
  return {
    checkOauthEnabled: executor(
      gql`
        query Auth_listOAuthProviders {
          auth_listOAuthProviders {
            enabled
            provider
          }
        }
      `,
      {
        transformer(data: AuthCheckOauthEnabledQuery) {
          return data.auth_listOAuthProviders;
        },
        vars(_: AuthCheckOauthEnabledQueryVariables) {},
      }
    ),
    addOauthCredientials: executor(
      gql`
        mutation Mutation($provider: String!, $state: String!, $code: String!) {
          oAuth_addLogin(provider: $provider, state: $state, code: $code)
        }
      `,
      {
        transformer(data: AuthAddOauthCredientialsMutation) {
          return data.oAuth_addLogin;
        },
        vars(_: AuthOauthLoginMutationVariables) {},
      }
    ),

    requestResetPassword: executor(
      gql`
        mutation Auth_requestResetPassword($email: String!) {
          auth_requestResetPassword(email: $email)
        }
      `,
      {
        transformer: (data: AuthRequestResetPasswordMutation) =>
          data.auth_requestResetPassword,
        vars(_: AuthRequestResetPasswordMutationVariables) {},
      }
    ),

    resetPassword: executor(
      gql`
        mutation Auth_requestResetPassword(
          $token: String!
          $password: String!
        ) {
          auth_resetPassword(token: $token, password: $password)
        }
      `,
      {
        transformer: (data: AuthResetPasswordMutation) =>
          data.auth_resetPassword,
        vars(_: AuthResetPasswordMutationVariables) {},
      }
    ),

    oauthLogin: executor(
      gql`
        mutation oAuth2($code: String!, $provider: String!, $state: String) {
          oAuth_login(code: $code, provider: $provider, state: $state) {
            id
          }
        }
      `,
      {
        transformer: (data: AuthOauthLoginMutation) => data.oAuth_login,

        vars(_: AuthOauthLoginMutationVariables) {},
      }
    ),

    verifyEmail: executor(
      gql`
        mutation VerifyEmail($token: String!) {
          auth_verifyEmail(token: $token) {
            id
          }
        }
      `,
      {
        transformer: (data: AuthVerifyEmailMutation) => data.auth_verifyEmail,
        vars(_: AuthVerifyEmailMutationVariables) {},
      }
    ),

    loginPageInitUrls: executor(
      gql`
        query Query {
          githubLoginUrl: oAuth_requestLogin(provider: "github")
          gitlabLoginUrl: oAuth_requestLogin(provider: "gitlab")
          googleLoginUrl: oAuth_requestLogin(provider: "google")
        }
      `,
      {
        transformer: (data: AuthLoginPageInitUrlsQuery) => data,
        vars(_: AuthLoginPageInitUrlsQueryVariables) {},
      }
    ),

    login: executor(
      gql`
        mutation Login($email: String!, $password: String!) {
          auth_login(email: $email, password: $password) {
            id
          }
        }
      `,
      {
        transformer: (data: AuthLoginMutation) => data.auth_login,
        vars(_: AuthLoginMutationVariables) {},
      }
    ),

    logout: executor(
      gql`
        mutation Auth {
          auth_logout
        }
      `,
      {
        transformer: (data: AuthLogoutMutation) => data.auth_logout,
        vars(_: AuthLogoutMutationVariables) {},
      }
    ),

    signUpWithEmail: executor(
      gql`
        mutation Auth_signup(
          $name: String!
          $password: String!
          $email: String!
        ) {
          auth_signup(name: $name, password: $password, email: $email) {
            id
          }
        }
      `,
      {
        transformer: (data: AuthSignUpWithEmailMutation) => data.auth_signup,
        vars(_: AuthSignUpWithEmailMutationVariables) {},
      }
    ),

    whoAmI: executor(
      gql`
        query Auth_me {
          auth_me {
            id
            email
            verified
          }
        }
      `,
      {
        transformer: (data: AuthWhoAmIQuery) => data.auth_me,
        vars(_: AuthWhoAmIQueryVariables) {},
      }
    ),
  };
};

export type AuthApiType = ReturnType<typeof GQLServerHandler>;
