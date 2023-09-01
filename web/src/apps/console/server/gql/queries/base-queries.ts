import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGqlReturn } from '~/root/lib/types/common';

export interface IGQLMethodsBase {
  accountCheckNameAvailability: (variables?: any) => IGqlReturn<any>;
  infraCheckNameAvailability: (variables?: any) => IGqlReturn<any>;
  coreCheckNameAvailability: (variables?: any) => IGqlReturn<any>;
  whoAmI: (variables?: any) => IGqlReturn<any>;
}

export const baseQueries = (executor: IExecutor): IGQLMethodsBase => ({
  accountCheckNameAvailability: executor(
    gql`
      query Query($name: String!) {
        accounts_checkNameAvailability(name: $name) {
          result
          suggestedNames
        }
      }
    `,
    { dataPath: 'accounts_checkNameAvailability' }
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
      dataPath: 'infra_checkNameAvailability',
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
      dataPath: 'core_checkNameAvailability',
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
});
