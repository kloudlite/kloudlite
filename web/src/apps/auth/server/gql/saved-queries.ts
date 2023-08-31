import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGQLServerHandler, IGqlReturn } from '~/root/lib/types/common';

export interface IGQLMethodsAuth {
  requestResetPassword: (variables?: { email: string }) => IGqlReturn<boolean>;
  resetPassword: (variables?: any) => IGqlReturn<any>;
  oauthLogin: (variables?: any) => IGqlReturn<any>;
  verifyEmail: (variables?: any) => IGqlReturn<any>;
  loginPageInitUrls: (variables?: any) => IGqlReturn<any>;
  login: (variables?: any) => IGqlReturn<any>;
  logout: (variables?: any) => IGqlReturn<any>;
  signUpWithEmail: (variables?: any) => IGqlReturn<any>;
  whoAmI: (variables?: any) => IGqlReturn<any>;
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
