import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '../helpers/execute-query-with-context';

export const GQLServerHandler = ({ headers }) => {
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
