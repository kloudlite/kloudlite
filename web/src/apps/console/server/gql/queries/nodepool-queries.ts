import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateNodePoolMutation,
  ConsoleCreateNodePoolMutationVariables,
  ConsoleGetNodePoolQuery,
  ConsoleGetNodePoolQueryVariables,
  ConsoleListNodePoolsQuery,
  ConsoleListNodePoolsQueryVariables,
} from '~/root/src/generated/gql/server';

export type INodepool = NN<ConsoleGetNodePoolQuery['infra_getNodePool']>;

export const nodepoolQueries = (executor: IExecutor) => ({
  getNodePool: executor(
    gql`
      query User($clusterName: String!, $poolName: String!) {
        infra_getNodePool(clusterName: $clusterName, poolName: $poolName) {
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
    `,
    {
      transformer(data: ConsoleGetNodePoolQuery) {
        return data.infra_getNodePool;
      },
      vars(_: ConsoleGetNodePoolQueryVariables) {},
    }
  ),
  createNodePool: executor(
    gql`
      mutation Infra_createNodePool($clusterName: String!, $pool: NodePoolIn!) {
        infra_createNodePool(clusterName: $clusterName, pool: $pool) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateNodePoolMutation) =>
        data.infra_createNodePool,
      vars(_: ConsoleCreateNodePoolMutationVariables) {},
    }
  ),
  listNodePools: executor(
    gql`
      query listNodePools(
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
      transformer: (data: ConsoleListNodePoolsQuery) =>
        data.infra_listNodePools,
      vars(_: ConsoleListNodePoolsQueryVariables) {},
    }
  ),
});
