import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const clusterQueries = (executor = ExecuteQueryWithContext({})) => ({
  createCluster: executor(gql`
    mutation Mutation($cluster: ClusterIn!) {
      infra_createCluster(cluster: $cluster) {
        id
      }
    }
  `),
  clustersCount: executor(
    gql`
      query Infra_listClusters {
        infra_listClusters {
          totalCount
        }
      }
    `,
    {
      dataPath: 'infra_listClusters',
    }
  ),

  listClusters: executor(
    gql`
      query Infra_listClusters {
        infra_listClusters {
          totalCount
          pageInfo {
            endCursor
            hasNextPage
            hasPreviousPage
            startCursor
          }
          edges {
            node {
              accountName
              apiVersion
              creationTime
              id
              kind
              markedForDeletion
              metadata {
                namespace
                name
                labels
                generation
                deletionTimestamp
                creationTimestamp
                annotations
              }
              recordVersion
              spec {
                vpc
                region
                operatorsHelmValuesRef {
                  namespace
                  key
                  name
                }
                nodeIps
                credentialsRef {
                  namespace
                  name
                }
                cloudProvider
                availabilityMode
                agentHelmValuesRef {
                  key
                  name
                  namespace
                }
                accountName
              }
            }
          }
        }
      }
    `,
    {
      dataPath: 'infra_listClusters',
    }
  ),
});
