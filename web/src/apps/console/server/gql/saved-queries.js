import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';
import { accountQueries } from './queries/account-queries';
import { projectQueries } from './queries/project-queries';
import { clusterQueries } from './queries/cluster-queries';

export const GQLServerHandler = ({ headers }) => {
  const executor = ExecuteQueryWithContext(headers);
  return {
    ...accountQueries(executor),
    ...projectQueries(executor),
    ...clusterQueries(executor),

    checkNameAvailability: executor(
      gql`
        query Infra_checkNameAvailability($resType: ResType!, $name: String!) {
          infra_checkNameAvailability(resType: $resType, name: $name) {
            suggestedNames
            result
          }
        }
      `,
      {
        dataPath: 'infra_checkNameAvailability',
      }
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
