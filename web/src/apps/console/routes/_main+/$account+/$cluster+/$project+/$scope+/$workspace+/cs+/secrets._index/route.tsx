import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import SubNavAction from '~/console/components/sub-nav-action';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import SecretResources from '~/console/page-components/secret-resource';
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

  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          secretsData: fake.ConsoleListProviderSecretsQuery
            .infra_listProviderSecrets as any,
        }}
      >
        {({ secretsData }) => {
          const secrets = parseNodes(secretsData);
          if (!secrets) {
            return null;
          }

          return (
            <>
              {secrets.length > 0 && (
                <SubNavAction deps={[secrets.length]}>
                  <Button
                    content="Add new config"
                    variant="primary"
                    onClick={() => {
                      setHandleSecret({ type: DIALOG_TYPE.ADD, data: null });
                    }}
                  />
                </SubNavAction>
              )}
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
                      setHandleSecret({ type: DIALOG_TYPE.ADD, data: null });
                    },
                  },
                }}
                tools={<Tools />}
              >
                <SecretResources items={secrets} linkComponent={Link} />
              </Wrapper>
            </>
          );
        }}
      </LoadingComp>
      <HandleSecret show={showHandleSecret} setShow={setHandleSecret} />
    </>
  );
};

export default Secrets;
