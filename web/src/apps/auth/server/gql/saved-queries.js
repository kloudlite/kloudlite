import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const GQLServerHandler = ({ headers }) => {
  const executor = ExecuteQueryWithContext(headers);
  return {
    loginPageInitUrls: executor(gql`
      query Query {
        githubLoginUrl: oAuth_requestLogin(provider: "github")
        gitlabLoginUrl: oAuth_requestLogin(provider: "gitlab")
        googleLoginUrl: oAuth_requestLogin(provider: "google")
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
