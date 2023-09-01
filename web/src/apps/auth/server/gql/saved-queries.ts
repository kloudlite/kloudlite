import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGQLServerHandler, IGqlReturn } from '~/root/lib/types/common';

export interface IGQLMethodsAuth {
  requestResetPassword: (variables: { email: string }) => IGqlReturn<boolean>;

  resetPassword: (variables: {
    token: string;
    password: string;
  }) => IGqlReturn<boolean>;

  oauthLogin: (variables: {
    code: string;
    provider: string;
    state?: string;
  }) => IGqlReturn<{ id: string }>;

  verifyEmail: (variables: { token: string }) => IGqlReturn<{ id: string }>;

  loginPageInitUrls: (variables?: any) => IGqlReturn<{
    githubLoginUrl?: string;
    gitlabLoginUrl?: string;
    googleLoginUrl?: string;
  }>;

  login: (variables: {
    email: string;
    password: string;
  }) => IGqlReturn<{ id: string }>;

  logout: (variables?: any) => IGqlReturn<boolean>;

  signUpWithEmail: (variables: {
    name: string;
    password: string;
    email: string;
  }) => IGqlReturn<{ id: string }>;

  whoAmI: (
    variables?: any
  ) => IGqlReturn<{ id: string; email: string; verified: boolean }>;
}

export const GQLServerHandler = ({
  headers,
  cookies,
}: IGQLServerHandler): IGQLMethodsAuth => {
  const executor = ExecuteQueryWithContext(headers, cookies);
  return {
    requestResetPassword: executor(gql`
      mutation Auth_requestResetPassword($email: String!) {
        auth_requestResetPassword(email: $email)
      }
    `),

    resetPassword: executor(gql`
      mutation Auth_requestResetPassword($token: String!, $password: String!) {
        auth_resetPassword(token: $token, password: $password)
      }
    `),

    oauthLogin: executor(
      gql`
        mutation oAuth2($code: String!, $provider: String!, $state: String) {
          oAuth_login(code: $code, provider: $provider, state: $state) {
            id
          }
        }
      `,
      {
        dataPath: 'oAuth_login',
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
      { dataPath: 'auth_verifyEmail' }
    ),

    loginPageInitUrls: executor(gql`
      query Query {
        githubLoginUrl: oAuth_requestLogin(provider: "github")
        gitlabLoginUrl: oAuth_requestLogin(provider: "gitlab")
        googleLoginUrl: oAuth_requestLogin(provider: "google")
      }
    `),

    login: executor(
      gql`
        mutation Login($email: String!, $password: String!) {
          auth_login(email: $email, password: $password) {
            id
          }
        }
      `,
      {
        dataPath: 'auth_login',
      }
    ),

    logout: executor(gql`
      mutation Auth {
        auth_logout
      }
    `),

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
      `
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
        dataPath: 'auth_me',
        transformer: (me) => ({ me }),
      }
    ),
  };
};
