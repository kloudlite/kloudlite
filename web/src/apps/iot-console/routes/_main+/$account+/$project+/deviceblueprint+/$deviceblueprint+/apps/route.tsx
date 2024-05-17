import { Plus } from '~/iotconsole/components/icons';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import {
  LoadingComp,
  pWrapper,
} from '~/iotconsole/components/loading-component';
import Wrapper from '~/iotconsole/components/wrapper';
import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import { parseNodes } from '~/iotconsole/server/r-utils/common';
import { ensureAccountSet } from '~/iotconsole/server/utils/auth-utils';
import { getPagination, getSearch } from '~/iotconsole/server/utils/common';
import { IRemixCtx } from '~/lib/types/common';
import { clearAppState } from '~/iotconsole/page-components/app-states';
import { useEffect } from 'react';
import { Button } from '~/components/atoms/button';
import Tools from './tools';
import AppResource from './app-resource';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  const { deviceblueprint } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listIotApps({
      deviceBlueprintName: deviceblueprint,
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      throw errors[0];
    }

    return { appsData: data };
  });

  return defer({ promise });
};

const Apps = () => {
  const { promise } = useLoaderData<typeof loader>();

  useEffect(() => {
    clearAppState();
  }, []);

  return (
    <LoadingComp
      data={promise}
      // skeletonData={{
      //   appsData: fake.ConsoleListAppsQuery.core_listApps as any,
      // }}
    >
      {({ appsData }) => {
        const apps = parseNodes(appsData);
        if (!apps) {
          return null;
        }

        return (
          <div>
            <Wrapper
              header={{
                title: 'Apps',
                action: apps?.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create new app"
                    prefix={<Plus />}
                    to="../new-app"
                    linkComponent={Link}
                  />
                ),
              }}
              empty={{
                is: apps?.length === 0,
                title: 'This is where you’ll manage your Apps.',
                content: (
                  <p>You can create a new app and manage the listed app.</p>
                ),
                action: {
                  content: 'Create new app',
                  prefix: <Plus />,
                  linkComponent: Link,
                  to: '../new-app',
                },
              }}
              tools={<Tools />}
            >
              <AppResource items={apps} />
            </Wrapper>
          </div>
        );
      }}
    </LoadingComp>
  );
};

export default Apps;
