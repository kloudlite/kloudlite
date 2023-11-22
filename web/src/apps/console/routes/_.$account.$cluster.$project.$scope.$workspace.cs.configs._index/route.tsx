import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import SubNavAction from '~/console/components/sub-nav-action';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import ConfigResources from '~/console/page-components/config-resource';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { DIALOG_TYPE } from '~/console/utils/commons';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import HandleConfig from './handle-config';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { project, scope, workspace } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listConfigs({
      project: {
        value: project,
        type: 'name',
      },
      scope: {
        value: workspace,
        type: scope === 'workspace' ? 'workspaceName' : 'environmentName',
      },
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      throw errors[0];
    }
    return { configsData: data };
  });

  return defer({ promise });
};

const Configs = () => {
  const [showHandleConfig, setHandleConfig] = useState<IShowDialog>(null);

  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          configsData: fake.ConsoleListConfigsQuery.core_listConfigs as any,
        }}
      >
        {({ configsData }) => {
          const configs = parseNodes(configsData);
          const { pageInfo, totalCount } = configsData;

          return (
            <>
              {configs.length > 0 && (
                <SubNavAction deps={[configs.length]}>
                  <Button
                    content="Add new config"
                    variant="primary"
                    onClick={() => {
                      setHandleConfig({ type: DIALOG_TYPE.ADD, data: null });
                    }}
                  />
                </SubNavAction>
              )}
              <Wrapper
                empty={{
                  is: configs.length === 0,
                  title: 'This is where you’ll manage your Config.',
                  content: (
                    <p>
                      You can create a new config and manage the listed configs.
                    </p>
                  ),
                  action: {
                    content: 'Create config',
                    prefix: <Plus />,
                    onClick: () => {
                      setHandleConfig({ type: DIALOG_TYPE.ADD, data: null });
                    },
                  },
                }}
                pagination={{
                  pageInfo,
                  totalCount,
                }}
                tools={<Tools />}
              >
                <ConfigResources items={configs} linkComponent={Link} />
              </Wrapper>
            </>
          );
        }}
      </LoadingComp>
      <HandleConfig show={showHandleConfig} setShow={setHandleConfig} />
    </>
  );
};

export default Configs;
