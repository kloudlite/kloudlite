import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateManagedServiceMutation,
  ConsoleCreateManagedServiceMutationVariables,
  ConsoleGetManagedServiceQuery,
  ConsoleGetManagedServiceQueryVariables,
  ConsoleGetTemplateQuery,
  ConsoleGetTemplateQueryVariables,
  ConsoleListManagedServicesQuery,
  ConsoleListManagedServicesQueryVariables,
  ConsoleListTemplatesQuery,
  ConsoleListTemplatesQueryVariables,
} from '~/root/src/generated/gql/server';

export type IManagedServiceTemplates =
  NN<ConsoleListTemplatesQuery>['core_listManagedServiceTemplates'];

export type IManagedServiceTemplate = NN<
  NN<ConsoleGetTemplateQuery>['core_getManagedServiceTemplate']
>;

export type IManagedService = NN<
  NN<ConsoleGetManagedServiceQuery>['core_getManagedService']
>;

export type IManagedServices = NN<
  NN<ConsoleListManagedServicesQuery>['core_listManagedServices']
>;

export const managedServiceQueries = (executor: IExecutor) => ({
  getTemplate: executor(
    gql`
      query Core_getManagedServiceTemplate($category: String!, $name: String!) {
        core_getManagedServiceTemplate(category: $category, name: $name) {
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
            kind
            name
            outputs {
              description
              label
              name
            }
          }
        }
      }
    `,
    {
      transformer(data: ConsoleGetTemplateQuery) {
        return data.core_getManagedServiceTemplate;
      },
      vars(_: ConsoleGetTemplateQueryVariables) {},
    }
  ),
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
  getManagedService: executor(
    gql`
      query Edges(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $name: String!
      ) {
        core_getManagedService(project: $project, scope: $scope, name: $name) {
          accountName
          apiVersion
          clusterName
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
          enabled
          id
          kind
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          markedForDeletion
          metadata {
            annotations
            creationTimestamp
            deletionTimestamp
            generation
            labels
            name
            namespace
          }
          recordVersion
          spec {
            serviceTemplate {
              apiVersion
              kind
              spec
            }
          }
          status {
            checks
            isReady
            lastReconcileTime
            message {
              RawMessage
            }
            resources {
              apiVersion
              kind
              name
              namespace
            }
          }
          syncStatus {
            action
            error
            lastSyncedAt
            recordVersion
            state
            syncScheduledAt
          }
          updateTime
        }
      }
    `,
    {
      transformer(data: ConsoleGetManagedServiceQuery) {
        return data.core_getManagedService;
      },
      vars(_: ConsoleGetManagedServiceQueryVariables) {},
    }
  ),
  listManagedServices: executor(
    gql`
      query Edges($project: ProjectId!, $scope: WorkspaceOrEnvId!) {
        core_listManagedServices(project: $project, scope: $scope) {
          edges {
            cursor
            node {
              accountName
              apiVersion
              clusterName
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              enabled
              id
              kind
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              markedForDeletion
              metadata {
                annotations
                creationTimestamp
                deletionTimestamp
                generation
                labels
                name
                namespace
              }
              recordVersion
              spec {
                serviceTemplate {
                  apiVersion
                  kind
                  spec
                }
              }
              status {
                checks
                isReady
                lastReconcileTime
                message {
                  RawMessage
                }
                resources {
                  apiVersion
                  kind
                  name
                  namespace
                }
              }
              syncStatus {
                action
                error
                lastSyncedAt
                recordVersion
                state
                syncScheduledAt
              }
              updateTime
            }
          }
          pageInfo {
            endCursor
            hasNextPage
            hasPreviousPage
            startCursor
          }
          totalCount
        }
      }
    `,
    {
      transformer(data: ConsoleListManagedServicesQuery) {
        return data.core_listManagedServices;
      },
      vars(_: ConsoleListManagedServicesQueryVariables) {},
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
