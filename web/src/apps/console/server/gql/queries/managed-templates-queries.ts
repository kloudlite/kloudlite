import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleGetMSvTemplateQuery,
  ConsoleGetMSvTemplateQueryVariables,
  ConsoleListMSvTemplatesQuery,
  ConsoleListMSvTemplatesQueryVariables,
} from '~/root/src/generated/gql/server';

export type IMSvTemplate = NN<
  ConsoleGetMSvTemplateQuery['infra_getManagedServiceTemplate']
>;
export type IMSvTemplates = NN<
  ConsoleListMSvTemplatesQuery['infra_listManagedServiceTemplates']
>;

export const managedTemplateQueries = (executor: IExecutor) => ({
  getMSvTemplate: executor(
    gql`
      query Infra_getManagedServiceTemplate(
        $category: String!
        $name: String!
      ) {
        infra_getManagedServiceTemplate(category: $category, name: $name) {
          active
          apiVersion
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
            displayUnit
            multiplier
          }
          kind
          logoUrl
          name
          outputs {
            description
            label
            name
          }
          resources {
            apiVersion
            description
            displayName
            kind
            name
          }
        }
      }
    `,
    {
      transformer(data: ConsoleGetMSvTemplateQuery) {
        return data.infra_getManagedServiceTemplate;
      },
      vars(_: ConsoleGetMSvTemplateQueryVariables) { },
    }
  ),
  listMSvTemplates: executor(
    gql`
      query Infra_listManagedServiceTemplates {
        infra_listManagedServiceTemplates {
          category
          displayName
          items {
            active
            apiVersion
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
              displayUnit
              multiplier
            }
            kind
            logoUrl
            name
            outputs {
              description
              label
              name
            }
            resources {
              apiVersion
              description
              displayName
              kind
              name
            }
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleListMSvTemplatesQuery) =>
        data.infra_listManagedServiceTemplates,
      vars(_: ConsoleListMSvTemplatesQueryVariables) { },
    }
  ),
});
