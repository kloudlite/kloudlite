import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateManagedServiceMutation,
  ConsoleCreateManagedServiceMutationVariables,
  ConsoleListTemplatesQuery,
  ConsoleListTemplatesQueryVariables,
} from '~/root/src/generated/gql/server';

export type IManagedServiceTemplates =
  NN<ConsoleListTemplatesQuery>['core_listManagedServiceTemplates'];

export const managedServiceQueries = (executor: IExecutor) => ({
  listTemplates: executor(
    gql`
      query Core_listManagedServiceTemplates {
        core_listManagedServiceTemplates {
          category
          displayName
          items {
            description
            active
            displayName
            fields {
              defaultValue
              inputType
              label
              max
              min
              name
              required
              unit
            }
            logoUrl
            name
            outputs {
              name
              label
              description
            }
            resources {
              description
              displayName
              fields {
                defaultValue
                inputType
                label
                max
                min
                name
                required
                unit
              }
              name
              outputs {
                description
                label
                name
              }
              kind
              apiVersion
            }
            kind
            apiVersion
          }
        }
      }
    `,
    {
      transformer(data: ConsoleListTemplatesQuery) {
        return data.core_listManagedServiceTemplates;
      },
      vars(_: ConsoleListTemplatesQueryVariables) {},
    }
  ),
  createManagedService: executor(
    gql`
      mutation Mutation($msvc: ManagedServiceIn!) {
        core_createManagedService(msvc: $msvc) {
          id
        }
      }
    `,
    {
      transformer(data: ConsoleCreateManagedServiceMutation) {
        return data.core_createManagedService;
      },
      vars(_: ConsoleCreateManagedServiceMutationVariables) {},
    }
  ),
});
