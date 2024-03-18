import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { clearAppState } from '~/console/page-components/app-states';
import { useEffect } from 'react';
import Tools from './tools';
import AppsResourcesV2 from './apps-resources-v2';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  const { environment, project } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listApps({
      envName: environment,
      projectName: project,
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
                    prefix={<PlusFill />}
                    to="../new-app"
                    LinkComponent={Link}
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
                  LinkComponent: Link,
                  to: '../new-app',
                },
              }}
              tools={<Tools />}
            >
              <AppsResourcesV2 items={apps} />
            </Wrapper>
          </div>
        );
      }}
    </LoadingComp>
  );
};

export default Apps;
