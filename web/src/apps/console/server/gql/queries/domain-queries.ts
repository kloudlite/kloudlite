import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
    ConsoleCreateDomainMutation,
    ConsoleCreateDomainMutationVariables,
    ConsoleDeleteDomainMutation,
    ConsoleDeleteDomainMutationVariables,
    ConsoleGetDomainQuery,
    ConsoleGetDomainQueryVariables,
    ConsoleListDomainsQuery,
    ConsoleListDomainsQueryVariables,
    ConsoleUpdateDomainMutation,
    ConsoleUpdateDomainMutationVariables,
} from '~/root/src/generated/gql/server';

export type IDomain = NN<ConsoleGetDomainQuery['infra_getDomainEntry']>;
export type IDomains = NN<ConsoleListDomainsQuery['infra_listDomainEntries']>;
export const domainQueries = (executor: IExecutor) => ({
  getDomain: executor(
    gql`
      query Infra_getDomainEntry($domainName: String!) {
        infra_getDomainEntry(domainName: $domainName) {
          updateTime
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          id
          domainName
          displayName
          creationTime
          createdBy {
            userEmail
            userId
            userName
          }
          clusterName
        }
      }
    `,
    {
      transformer: (data: ConsoleGetDomainQuery) => data.infra_getDomainEntry,
      vars(_: ConsoleGetDomainQueryVariables) {},
    }
  ),
  createDomain: executor(
    gql`
      mutation Mutation($domainEntry: DomainEntryIn!) {
        infra_createDomainEntry(domainEntry: $domainEntry) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateDomainMutation) =>
        data.infra_createDomainEntry,
      vars(_: ConsoleCreateDomainMutationVariables) {},
    }
  ),
  updateDomain: executor(
    gql`
      mutation Infra_updateDomainEntry($domainEntry: DomainEntryIn!) {
        infra_updateDomainEntry(domainEntry: $domainEntry) {
          id
        }
      }
    `,
    {
      transformer(data: ConsoleUpdateDomainMutation) {
        return data.infra_updateDomainEntry;
      },
      vars(_: ConsoleUpdateDomainMutationVariables) {},
    }
  ),
  deleteDomain: executor(
    gql`
      mutation Infra_deleteDomainEntry($domainName: String!) {
        infra_deleteDomainEntry(domainName: $domainName)
      }
    `,
    {
      transformer(data: ConsoleDeleteDomainMutation) {
        return data.infra_deleteDomainEntry;
      },
      vars(_: ConsoleDeleteDomainMutationVariables) {},
    }
  ),
  listDomains: executor(
    gql`
      query Infra_listDomainEntries(
        $search: SearchDomainEntry
        $pagination: CursorPaginationIn
      ) {
        infra_listDomainEntries(search: $search, pagination: $pagination) {
          pageInfo {
            endCursor
            hasNextPage
            hasPreviousPage
            startCursor
          }
          totalCount
          edges {
            cursor
            node {
              updateTime
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              id
              domainName
              displayName
              creationTime
              createdBy {
                userEmail
                userId
                userName
              }
            }
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleListDomainsQuery) =>
        data.infra_listDomainEntries,
      vars(_: ConsoleListDomainsQueryVariables) {},
    }
  ),
});
