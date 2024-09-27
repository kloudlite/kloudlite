import { Plus } from '~/console/components/icons';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/lib/types/common';
import { clearAppState } from '~/console/page-components/app-states';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import fake from '~/root/fake-data-generator/fake';
import { EmptyAppImage } from '~/console/components/empty-resource-images';
import Tools from './tools';
import ExternalNameResource from './external-app-resource';
import HandleExternalApp from './handle-external-app';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  const { environment } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listExternalApps({
      envName: environment,
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      throw errors[0];
    }

    return { externalAppsData: data };
  });

  return defer({ promise });
};

const ExternalApp = () => {
  const [visible, setVisible] = useState(false);

  const { promise } = useLoaderData<typeof loader>();
  useEffect(() => {
    clearAppState();
  }, []);

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          externalAppsData: fake.ConsoleListExternalAppsQuery
            .core_listExternalApps as any,
        }}
      >
        {({ externalAppsData }) => {
          const externalApps = parseNodes(externalAppsData);
          if (!externalApps) {
            return null;
          }

          return (
            <div>
              <Wrapper
                header={{
                  title: 'External App',
                  action: externalApps?.length > 0 && (
                    <Button
                      variant="primary"
                      content="Create external app"
                      prefix={<Plus />}
                      onClick={() => {
                        setVisible(true);
                      }}
                      // to="../new-app"
                      // linkComponent={Link}
                    />
                  ),
                }}
                empty={{
                  image: <EmptyAppImage />,
                  is: externalApps?.length === 0,
                  title: 'This is where you’ll manage your external Apps.',
                  content: (
                    <p>
                      You can create a new app and manage the listed external
                      app.
                    </p>
                  ),
                  action: {
                    content: 'Create external app',
                    prefix: <Plus />,
                    onClick: () => {
                      setVisible(true);
                    },
                    linkComponent: Link,
                    // to: '../new-app',
                  },
                }}
                tools={<Tools />}
              >
                <ExternalNameResource items={externalApps} />
              </Wrapper>
            </div>
          );
        }}
      </LoadingComp>
      <HandleExternalApp
        {...{
          visible,
          setVisible,
          isUpdate: false,
        }}
      />
    </>
  );
};

export default ExternalApp;
