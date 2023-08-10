import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const nodepoolQueries = (executor = ExecuteQueryWithContext({})) => ({
  createNodePool: executor(
    gql`
      mutation Infra_createNodePool($clusterName: String!, $pool: NodePoolIn!) {
        infra_createNodePool(clusterName: $clusterName, pool: $pool) {
          id
        }
      }
    `,
    {
      dataPath: 'infra_createNodePool',
    }
  ),
  listNodePools: executor(
    gql`
      query Edges(
        $clusterName: String!
        $search: SearchFilter
        $pagination: PaginationQueryArgs
      ) {
        infra_listNodePools(
          clusterName: $clusterName
          search: $search
          pagination: $pagination
        ) {
          edges {
            node {
              updateTime
              spec {
                targetCount
                minCount
                maxCount
                awsNodeConfig {
                  vpc
                  spotSpecs {
                    memMin
                    memMax
                    cpuMin
                    cpuMax
                  }
                  region
                  provisionMode
                  onDemandSpecs {
                    instanceType
                  }
                  isGpu
                  imageId
                }
              }
              metadata {
                name
                annotations
              }
              clusterName
              status {
                isReady
                message {
                  RawMessage
                }
                checks
              }
            }
          }
          pageInfo {
            startCursor
            hasPreviousPage
            hasNextPage
            endCursor
          }
          totalCount
        }
      }
    `,
    {
      dataPath: 'infra_listNodePools',
    }
  ),
});
