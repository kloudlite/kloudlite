import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleListNamespacesQuery,
  ConsoleListNamespacesQueryVariables,
} from '~/root/src/generated/gql/server';

export type INamespaces = NN<
  ConsoleListNamespacesQuery['infra_listNamespaces']
>;

export const namespaceQueries = (executor: IExecutor) => ({
  listNamespaces: executor(
    gql`
      query Infra_listNamespaces($clusterName: String!) {
        infra_listNamespaces(clusterName: $clusterName) {
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
                finalizers
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
      transformer: (data: ConsoleListNamespacesQuery) =>
        data.infra_listNamespaces,
      vars(_: ConsoleListNamespacesQueryVariables) {},
    }
  ),
});
