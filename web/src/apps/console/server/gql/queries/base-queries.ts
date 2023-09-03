import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';

export const baseQueries = (executor: IExecutor) => ({
  accountCheckNameAvailability: executor(
    gql`
      query Query($name: String!) {
        accounts_checkNameAvailability(name: $name) {
          result
          suggestedNames
        }
      }
    `,
    {
      transformer(data) {},
    }
  ),

  infraCheckNameAvailability: executor(
    gql`
      query Infra_checkNameAvailability($resType: ResType!, $name: String!) {
        infra_checkNameAvailability(resType: $resType, name: $name) {
          suggestedNames
          result
        }
      }
    `,
    {
      transformer(data) {},
    }
  ),

  coreCheckNameAvailability: executor(
    gql`
      query Core_checkNameAvailability(
        $resType: ConsoleResType!
        $name: String!
        $namespace: String
      ) {
        core_checkNameAvailability(
          resType: $resType
          name: $name
          namespace: $namespace
        ) {
          result
          suggestedNames
        }
      }
    `,
    {
      transformer(data) {},
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
      transformer: (me) => ({ me }),
    }
  ),
});
