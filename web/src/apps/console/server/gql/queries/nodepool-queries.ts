import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGqlReturn } from '~/root/lib/types/common';

export interface IGQLMethodsNodepool {
  createNodePool: (variables?: any) => IGqlReturn<any>;
  listNodePools: (variables?: any) => IGqlReturn<any>;
}

export const nodepoolQueries = (executor: IExecutor): IGQLMethodsNodepool => ({
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
        $pagination: CursorPaginationIn
        $search: SearchNodepool
      ) {
        infra_listNodePools(
          clusterName: $clusterName
          pagination: $pagination
          search: $search
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
