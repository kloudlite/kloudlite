import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  ConsoleAccountCheckNameAvailabilityQuery,
  ConsoleAccountCheckNameAvailabilityQueryVariables,
  ConsoleCoreCheckNameAvailabilityQuery,
  ConsoleCoreCheckNameAvailabilityQueryVariables,
  ConsoleCrCheckNameAvailabilityQuery,
  ConsoleCrCheckNameAvailabilityQueryVariables,
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
  crCheckNameAvailability: executor(
    gql`
      query CR_checkUserNameAvailability($name: String!) {
        cr_checkUserNameAvailability(name: $name) {
          result
          suggestedNames
        }
      }
    `,
    {
      transformer: (data: ConsoleCrCheckNameAvailabilityQuery) =>
        data.cr_checkUserNameAvailability,
      vars(_: ConsoleCrCheckNameAvailabilityQueryVariables) {},
    }
  ),
  infraCheckNameAvailability: executor(
    gql`
      query Infra_checkNameAvailability(
        $resType: ResType!
        $name: String!
        $clusterName: String
      ) {
        infra_checkNameAvailability(
          resType: $resType
          name: $name
          clusterName: $clusterName
        ) {
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
          providerGitlab
          providerGithub
          providerGoogle
        }
      }
    `,
    {
      transformer: (data: ConsoleWhoAmIQuery) => data.auth_me,
      vars(_: ConsoleWhoAmIQueryVariables) {},
    }
  ),
});
