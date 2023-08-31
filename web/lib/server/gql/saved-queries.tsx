import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '../helpers/execute-query-with-context';
import { IGQLServerHandler, IGqlReturn } from '../../types/common';

interface GQLServerHandlerReturn {
  whoAmI: (variables?: any) => IGqlReturn<{ me: any }>;
}

export const GQLServerHandler = ({
  headers,
}: IGQLServerHandler): GQLServerHandlerReturn => {
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
