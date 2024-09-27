import { defer } from '@remix-run/node';
import { Link, useLoaderData, useParams } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { Plus } from '~/console/components/icons';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import HandleScope from '~/console/page-components/handle-environment';
import { parseNodes } from '~/console/server/r-utils/common';
import { getPagination, getSearch } from '~/console/server/utils/common';
import fake from '~/root/fake-data-generator/fake';
import { IRemixCtx } from '~/root/lib/types/common';

import { EmptyEnvironmentImage } from '~/console/components/empty-resource-images';
import { IEnvironment } from '~/console/server/gql/queries/environment-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import EnvironmentResourcesV2 from './environment-resources-v2';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listEnvironments({
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });

    const { data: clusterData, errors: clusterErrors } = await GQLServerHandler(
      ctx.request
    ).listAllClusters({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });

    if (clusterErrors) {
      throw clusterErrors[0];
    }

    if (errors) {
      throw errors[0];
    }

    return {
      environmentData: data || {},
      clusterList: clusterData || {},
    };
  });

  return defer({ promise });
};

const Workspaces = () => {
  const [showAddWS, setShowAddWS] =
    useState<IShowDialog<IEnvironment | null>>(null);

  const { promise } = useLoaderData<typeof loader>();
  const { account } = useParams();

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          environmentData: fake.ConsoleListEnvironmentsQuery
            .core_listEnvironments as any,
          clusterList: fake.ConsoleListAllClustersQuery.byok_clusters as any,
        }}
      >
        {({ environmentData, clusterList }) => {
          const environments = parseNodes(environmentData);
          const clusters = parseNodes(clusterList);

          if (!environments) {
            return null;
          }

          if (clusters?.length === 0) {
            return (
              <Wrapper
                header={{
                  title: 'Environments',
                }}
                empty={{
                  image: <EmptyEnvironmentImage />,
                  is: environments?.length === 0,
                  title: 'This is where you’ll manage your environment.',
                  content: (
                    <p>
                      You don't have any compute attached to your account.
                      Please attach a compute to your account to create an
                      environment.
                      <br />
                      Go to{' '}
                      <Link
                        to={`/${account}/infra/clusters`}
                        className="text-text-default"
                      >
                        <span className="bodyMd-semibold underline underline-offset-1 text-text-default">
                          Infrastructure
                        </span>
                      </Link>{' '}
                      to attach your compute or local device.
                    </p>
                    /* <Button
                        size="sm"
                        content={
                          <span className="truncate text-left">
                            Infrastructure
                          </span>
                        }
                        variant="primary-plain"
                        className="truncate justify-center"
                        to={`/${account}/infra/clusters`}
                      /> */
                  ),
                }}
                tools={<Tools />}
                pagination={environmentData}
              >
                <EnvironmentResourcesV2 items={environments || []} />
              </Wrapper>
            );
          }

          return (
            <Wrapper
              header={{
                title: 'Environments',
                action: environments.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create Environment"
                    prefix={<Plus />}
                    onClick={() => {
                      setShowAddWS({ type: 'add', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                image: <EmptyEnvironmentImage />,
                is: environments?.length === 0,
                title: 'This is where you’ll manage your environment.',
                content: (
                  <p>
                    You can create a new workspace and manage the listed
                    workspaces.
                  </p>
                ),
                action: {
                  content: 'Create new environment',
                  prefix: <Plus />,
                  onClick: () => {
                    setShowAddWS({ type: 'add', data: null });
                  },
                },
              }}
              tools={<Tools />}
              pagination={environmentData}
            >
              <EnvironmentResourcesV2 items={environments || []} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleScope show={showAddWS} setShow={setShowAddWS} />
    </>
  );
};
export default Workspaces;
