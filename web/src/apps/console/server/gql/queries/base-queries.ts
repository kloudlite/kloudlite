import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  ConsoleAccountCheckNameAvailabilityQuery,
  ConsoleAccountCheckNameAvailabilityQueryVariables,
  ConsoleCoreCheckNameAvailabilityQuery,
  ConsoleCoreCheckNameAvailabilityQueryVariables,
  ConsoleInfraCheckNameAvailabilityQuery,
  ConsoleInfraCheckNameAvailabilityQueryVariables,
  ConsoleWhoAmIQuery,
  ConsoleWhoAmIQueryVariables,
} from '~/root/src/generated/gql/server';

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
      transformer: (data: ConsoleAccountCheckNameAvailabilityQuery) =>
        data.accounts_checkNameAvailability,
      vars(_: ConsoleAccountCheckNameAvailabilityQueryVariables) {},
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
      transformer: (data: ConsoleInfraCheckNameAvailabilityQuery) =>
        data.infra_checkNameAvailability,
      vars(_: ConsoleInfraCheckNameAvailabilityQueryVariables) {},
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
      transformer: (data: ConsoleCoreCheckNameAvailabilityQuery) =>
        data.core_checkNameAvailability,
      vars(_: ConsoleCoreCheckNameAvailabilityQueryVariables) {},
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
      transformer: (data: ConsoleWhoAmIQuery) => data.auth_me,
      vars(_: ConsoleWhoAmIQueryVariables) {},
    }
  ),
});
