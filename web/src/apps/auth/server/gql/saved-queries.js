import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const GQLServerHandler = ({ headers, cookies }) => {
  const executor = ExecuteQueryWithContext(headers, cookies);
  return {
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
