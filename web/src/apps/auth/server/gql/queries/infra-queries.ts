import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';

export const infraQueries = (executor: IExecutor) => ({
  cli_CoreCheckNameAvailability: executor(
    gql`
      query Core_checkNameAvailability(
        $resType: ConsoleResType!
        $name: String!
      ) {
        core_checkNameAvailability(resType: $resType, name: $name) {
          result
          suggestedNames
        }
      }
    `,
    {
      transformer: (data: any) => data.core_checkNameAvailability,
      vars: (_: any) => {},
    }
  ),
});
