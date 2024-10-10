import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleDeleteRegistryImageMutation,
  ConsoleDeleteRegistryImageMutationVariables,
  ConsoleGetRegistryImageQuery,
  ConsoleGetRegistryImageQueryVariables,
  ConsoleGetRegistryImageUrlQuery,
  ConsoleGetRegistryImageUrlQueryVariables,
  ConsoleListRegistryImagesQuery,
  ConsoleListRegistryImagesQueryVariables,
  ConsoleSearchRegistryImagesQuery,
  ConsoleSearchRegistryImagesQueryVariables,
} from '~/root/src/generated/gql/server';

export type IRegistryImages = NN<
  ConsoleListRegistryImagesQuery['core_listRegistryImages']
>;
export type IRegistryImage = NN<
  ConsoleGetRegistryImageQuery['core_getRegistryImage']
>;

export const registryImagesQueries = (executor: IExecutor) => ({
  deleteRegistryImage: executor(
    gql`
      mutation Core_deleteRegistryImage($image: String!) {
        core_deleteRegistryImage(image: $image)
      }
    `,
    {
      transformer: (data: ConsoleDeleteRegistryImageMutation) =>
        data.core_deleteRegistryImage,
      vars(_: ConsoleDeleteRegistryImageMutationVariables) {},
    }
  ),
  getRegistryImage: executor(
    gql`
      query Core_getRegistryImage($image: String!) {
        core_getRegistryImage(image: $image) {
          accountName
          creationTime
          id
          imageName
          imageTag
          markedForDeletion
          meta
          recordVersion
          updateTime
        }
      }
    `,
    {
      transformer: (data: ConsoleGetRegistryImageQuery) =>
        data.core_getRegistryImage,
      vars(_: ConsoleGetRegistryImageQueryVariables) {},
    }
  ),
  getRegistryImageUrl: executor(
    gql`
      query Core_getRegistryImageURL {
        core_getRegistryImageURL {
          scriptUrl
          url
          urlExample
          scriptUrlExample
          klWebhookAuthToken
        }
      }
    `,
    {
      transformer: (data: ConsoleGetRegistryImageUrlQuery) =>
        data.core_getRegistryImageURL,
      vars(_: ConsoleGetRegistryImageUrlQueryVariables) {},
    }
  ),
  searchRegistryImages: executor(
    gql`
      query Core_searchRegistryImages($query: String!) {
        core_searchRegistryImages(query: $query) {
          accountName
          creationTime
          id
          imageName
          imageTag
          markedForDeletion
          meta
          recordVersion
          updateTime
        }
      }
    `,
    {
      transformer: (data: ConsoleSearchRegistryImagesQuery) => {
        return data.core_searchRegistryImages;
      },
      vars(_: ConsoleSearchRegistryImagesQueryVariables) {},
    }
  ),
  listRegistryImages: executor(
    gql`
      query Core_listRegistryImages($pq: CursorPaginationIn) {
        core_listRegistryImages(pq: $pq) {
          edges {
            cursor
            node {
              accountName
              creationTime
              id
              imageName
              imageTag
              markedForDeletion
              meta
              recordVersion
              updateTime
            }
          }
          pageInfo {
            endCursor
            hasNextPage
            hasPrevPage
            startCursor
          }
          totalCount
        }
      }
    `,
    {
      transformer: (data: ConsoleListRegistryImagesQuery) => {
        return data.core_listRegistryImages;
      },
      vars(_: ConsoleListRegistryImagesQueryVariables) {},
    }
  ),
});
