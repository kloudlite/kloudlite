import { useState } from 'react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import Wrapper from '~/console/components/wrapper';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { useLoaderData, Link } from '@remix-run/react';
import { defer } from '@remix-run/node';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { IRemixCtx } from '~/root/lib/types/common';
import {
  getPagination,
  getScopeAndProjectQuery,
  getSearch,
} from '~/console/server/utils/common';
import { parseNodes } from '~/console/server/r-urils/common';
import { dummyData } from '~/console/dummy/data';
import Tools from './tools';
import AppsResources from './apps-resources';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listApps({
      ...getScopeAndProjectQuery(ctx),
      pagination: getPagination(ctx),
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
  const [viewMode, setViewMode] = useState<'list' | 'grid'>('list');

  const { promise } = useLoaderData<typeof loader>();
  console.log('promise', promise);

  return (
    <LoadingComp data={promise}>
      {({ appsData }) => {
        console.log(appsData);
        const apps = parseNodes(appsData);
        apps.length = 1;
        if (!apps) {
          return null;
        }
        return (
          <Wrapper
            header={{
              title: 'Apps',
              action: apps.length > 0 && (
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
              is: apps.length === 0,
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
          >
            <Tools viewMode={viewMode} setViewMode={setViewMode} />
            <AppsResources items={dummyData.apps} />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default Apps;
