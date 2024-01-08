import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleListHelmChartQuery,
  ConsoleGetHelmChartQueryVariables,
  ConsoleListHelmChartQueryVariables,
  ConsoleGetHelmChartQuery,
  ConsoleCreateHelmChartMutation,
  ConsoleCreateHelmChartMutationVariables,
  ConsoleUpdateHelmChartMutation,
  ConsoleUpdateHelmChartMutationVariables,
  ConsoleDeleteHelmChartMutation,
  ConsoleDeleteHelmChartMutationVariables,
} from '~/root/src/generated/gql/server';

export type IHelmCharts = NN<
  ConsoleListHelmChartQuery['infra_listHelmReleases']
>;

export const helmChartQueries = (executor: IExecutor) => ({
  getHelmChart: executor(
    gql`
      query Infra_getHelmRelease($clusterName: String!, $name: String!) {
        infra_getHelmRelease(clusterName: $clusterName, name: $name) {
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          markedForDeletion
          metadata {
            name
            namespace
          }
          spec {
            chartName
            chartRepoURL
            chartVersion
            values
          }
          status {
            checks
            isReady
            lastReadyGeneration
            lastReconcileTime
            message {
              RawMessage
            }
            releaseNotes
            releaseStatus
            resources {
              apiVersion
              kind
              name
              namespace
            }
          }
          updateTime
        }
      }
    `,
    {
      transformer(data: ConsoleGetHelmChartQuery) {
        return data.infra_getHelmRelease;
      },
      vars(_: ConsoleGetHelmChartQueryVariables) { },
    }
  ),
  listHelmChart: executor(
    gql`
      query Infra_listHelmReleases($clusterName: String!) {
        infra_listHelmReleases(clusterName: $clusterName) {
          totalCount
          edges {
            node {
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              markedForDeletion
              metadata {
                name
                namespace
              }
              spec {
                chartName
                chartRepoURL
                chartVersion
                values
              }
              status {
                checks
                isReady
                lastReadyGeneration
                lastReconcileTime
                message {
                  RawMessage
                }
                releaseNotes
                releaseStatus
                resources {
                  apiVersion
                  kind
                  name
                  namespace
                }
              }
              updateTime
            }
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleListHelmChartQuery) =>
        data.infra_listHelmReleases,
      vars(_: ConsoleListHelmChartQueryVariables) { },
    }
  ),
  createHelmChart: executor(
    gql`
      mutation Infra_createHelmRelease(
        $clusterName: String!
        $release: HelmReleaseIn!
      ) {
        infra_createHelmRelease(clusterName: $clusterName, release: $release) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateHelmChartMutation) =>
        data.infra_createHelmRelease,
      vars(_: ConsoleCreateHelmChartMutationVariables) { },
    }
  ),
  updateHelmChart: executor(
    gql`
      mutation Infra_updateHelmRelease(
        $clusterName: String!
        $release: HelmReleaseIn!
      ) {
        infra_updateHelmRelease(clusterName: $clusterName, release: $release) {
          id
        }
      }
    `,
    {
      transformer(data: ConsoleUpdateHelmChartMutation) {
        return data.infra_updateHelmRelease;
      },
      vars(_: ConsoleUpdateHelmChartMutationVariables) { },
    }
  ),
  deleteHelmChart: executor(
    gql`
      mutation Infra_deleteHelmRelease(
        $clusterName: String!
        $releaseName: String!
      ) {
        infra_deleteHelmRelease(
          clusterName: $clusterName
          releaseName: $releaseName
        )
      }
    `,
    {
      transformer(data: ConsoleDeleteHelmChartMutation) {
        return data.infra_deleteHelmRelease;
      },
      vars(_: ConsoleDeleteHelmChartMutationVariables) { },
    }
  ),
});
