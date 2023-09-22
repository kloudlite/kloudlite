import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import SecretResource from '~/console/page-components/secret-resource';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import { IRemixCtx } from '~/root/lib/types/common';
import HandleSecret from './handle-secret';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { project, scope, workspace } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listSecrets({
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
    return { secretsData: data };
  });

  return defer({ promise });
};

const Secrets = () => {
  const [showHandleSecret, setHandleSecret] = useState<IShowDialog>(null);

  const { setData: setSubNavAction } = useSubNavData();

  useEffect(() => {
    setSubNavAction({
      show: true,
      content: 'Add new secret',
      action: () => {
        setHandleSecret({ type: 'add', data: null });
      },
    });
  }, []);

  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp data={promise}>
        {({ secretsData }) => {
          const secrets = parseNodes(secretsData);
          if (!secrets) {
            return null;
          }
          return (
            <Wrapper
              empty={{
                is: secrets.length === 0,
                title: 'This is where you’ll manage your Secret.',
                content: (
                  <p>
                    You can create a new secret and manage the listed secrets.
                  </p>
                ),
                action: {
                  content: 'Create secret',
                  prefix: <Plus />,
                  LinkComponent: Link,
                  onClick: () => {
                    setHandleSecret({ type: 'add', data: null });
                  },
                },
              }}
              tools={<Tools />}
            >
              <SecretResource
                onDelete={() => {}}
                items={secrets}
                linkComponent={Link}
              />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleSecret show={showHandleSecret} setShow={setHandleSecret} />
    </>
  );
};

export default Secrets;
