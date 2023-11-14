import gql from 'graphql-tag';
import {
  LibWhoAmIQuery,
  LibWhoAmIQueryVariables,
} from '~/root/src/generated/gql/server';
import { IGQLServerProps, NN } from '../../types/common';
import { ExecuteQueryWithContext } from '../helpers/execute-query-with-context';

export type UserMe = NN<LibWhoAmIQuery['auth_me']>;

export const GQLServerHandler = ({ headers }: IGQLServerProps) => {
  const executor = ExecuteQueryWithContext(headers);
  return {
    whoAmI: executor(
      gql`
        query Me {
          auth_me {
            verified
            name
            id
            email
            providerGitlab
            providerGithub
            providerGoogle
          }
        }
      `,
      {
        transformer: (data: LibWhoAmIQuery) => data.auth_me,
        vars(_: LibWhoAmIQueryVariables) {},
      }
    ),
  };
};

export type LibApiType = ReturnType<typeof GQLServerHandler>;
