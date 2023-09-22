import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import ConfigResource from '~/console/page-components/config-resource';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import { IRemixCtx } from '~/root/lib/types/common';
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
      pagination: getPagination(ctx),
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
  const { setData: setSubNavAction } = useSubNavData();

  useEffect(() => {
    setSubNavAction({
      show: true,
      content: 'Add new config',
      action: () => {
        setHandleConfig({ type: 'add', data: null });
      },
    });
  }, []);

  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp data={promise}>
        {({ configsData }) => {
          const configs = parseNodes(configsData);
          return (
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
                    setHandleConfig({ type: 'add', data: null });
                  },
                },
              }}
              tools={<Tools />}
            >
              <ConfigResource
                onDelete={() => {}}
                items={configs}
                linkComponent={Link}
              />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleConfig show={showHandleConfig} setShow={setHandleConfig} />
    </>
  );
};

export default Configs;
