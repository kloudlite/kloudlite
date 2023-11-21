import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button.jsx';
import Wrapper from '~/console/components/wrapper';
import { parseNodes } from '~/console/server/r-utils/common';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { MainLayoutSK } from '~/console/page-components/skeletons';
import { LoadingComp, pWrapper } from '../../components/loading-component';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import { ensureAccountSet } from '../../server/utils/auth-utils';
import ClusterResources from './cluster-resources';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(ctx.request).listClusters({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });

    if (errors) {
      throw errors[0];
    }

    if (data.edges.length === 0) {
      const { data: secrets, errors: sErrors } = await GQLServerHandler(
        ctx.request
      ).listProviderSecrets({});

      if (sErrors) {
        throw sErrors[0];
      }

      return {
        clustersData: data || {},
        secretsCount: secrets.edges.length,
      };
    }

    return {
      clustersData: data,
      secretsCount: -1,
    };
  });

  return defer({ promise });
};

const Clusters = () => {
  const { promise } = useLoaderData<typeof loader>();

  const { account } = useParams();

  const getEmptyState = ({
    clustersCount,
    cloudProviderSecretsCount,
  }: {
    clustersCount: number;
    cloudProviderSecretsCount: number;
  }) => {
    if (cloudProviderSecretsCount === 0) {
      return {
        is: true,
        title: 'please setup your cloud provider first',
        content: (
          <p>
            you need to setup your add at least one cloud provider first, before
            starting working with clusters
          </p>
        ),
        action: {
          content: 'Setup Cloud Provider and Cluster',
          prefix: <Plus />,
          LinkComponent: Link,
          to: `/onboarding/${account}/new-cloud-provider`,
        },
      };
    }

    if (clustersCount === 0) {
      return {
        is: true,
        title: 'This is where you’ll manage your cluster.',
        content: (
          <p>You can create a new cluster and manage the listed cluster.</p>
        ),
        action: {
          content: 'Create new cluster',
          prefix: <Plus />,
          LinkComponent: Link,
          to: `/${account}/new-cluster`,
        },
      };
    }

    return {
      is: false,
      title: 'This is where you’ll manage your cluster.',
      content: (
        <p>You can create a new cluster and manage the listed cluster.</p>
      ),
      action: {
        content: 'Create new cluster',
        prefix: <Plus />,
        LinkComponent: Link,
        to: `/${account}/new-cluster`,
      },
    };
  };

  return (
    <LoadingComp data={promise} skeleton={<MainLayoutSK title="Clusters" />}>
      {({ clustersData, secretsCount }) => {
        const clusters = parseNodes(clustersData);

        if (!clusters) {
          return null;
        }

        const { pageInfo, totalCount } = clustersData;
        return (
          <Wrapper
            header={{
              title: 'Cluster',
              action: clusters.length > 0 && (
                <Button
                  variant="primary"
                  content="Create Cluster"
                  prefix={<PlusFill />}
                  LinkComponent={Link}
                  to={`/${account}/new-cluster`}
                />
              ),
            }}
            empty={getEmptyState({
              clustersCount: clusters.length,
              cloudProviderSecretsCount: secretsCount,
            })}
            pagination={{
              pageInfo,
              totalCount,
            }}
            tools={<Tools />}
          >
            <ClusterResources items={clusters} />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default Clusters;
