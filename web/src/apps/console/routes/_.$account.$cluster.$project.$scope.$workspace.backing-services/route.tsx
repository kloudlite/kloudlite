import { Plus, PlusFill } from '@jengaicons/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import { defer } from '@remix-run/node';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { IRemixCtx } from '~/root/lib/types/common';
import { useLoaderData } from '@remix-run/react';
import BackendServicesResources from './backend-services-resources';
import HandleBackendService from './handle-backend-service';
import Tools from './tools';

export const loader = (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listTemplates(
      {}
    );
    if (errors) {
      throw errors[0];
    }
    return { templates: data };
  });
  return defer({ promise });
};

const BackendServices = () => {
  const [showHandleBackendService, setShowHanldeBackendService] =
    useState<IShowDialog>(null);
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp data={promise}>
      {({ templates }) => {
        const backendServices = [
          {
            name: 'MongoDb',
            id: 'mongodb',
            type: 'External',
            updateInfo: {
              author: 'Bikash updated the service',
              time: '3 days ago',
            },
          },
          {
            name: 'AWS',
            id: 'aws',
            type: 'External',
            updateInfo: {
              author: 'Bikash updated the service',
              time: '3 days ago',
            },
          },
          {
            name: 'Postgres',
            id: 'postgres',
            type: 'External',
            updateInfo: {
              author: 'Bikash updated the service',
              time: '3 days ago',
            },
          },
          {
            name: 'Redis',
            id: 'redis',
            type: 'External',
            updateInfo: {
              author: 'Bikash updated the service',
              time: '3 days ago',
            },
          },
          {
            name: 'Kafka',
            id: 'kafka',
            type: 'External',
            updateInfo: {
              author: 'Bikash updated the service',
              time: '3 days ago',
            },
          },
          {
            name: 'Postgres',
            id: 'postgres-1',
            type: 'External',
            updateInfo: {
              author: 'Bikash updated the service',
              time: '3 days ago',
            },
          },
        ];
        backendServices.length = 0;
        return (
          <>
            <Wrapper
              header={{
                title: 'Backing services',
                action: backendServices.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create new app"
                    prefix={<PlusFill />}
                    onClick={() => {
                      setShowHanldeBackendService({ type: 'add', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                is: backendServices.length === 0,
                title: 'This is where you’ll manage your Backing services.',
                content: (
                  <p>
                    You can create a new backing service and manage the listed
                    backing service.
                  </p>
                ),
                action: {
                  content: 'Create new backing service',
                  prefix: <Plus />,
                  onClick: () => {
                    setShowHanldeBackendService({ type: 'add', data: null });
                    console.log('open');
                  },
                },
              }}
            >
              <Tools />
              <BackendServicesResources items={backendServices} />
            </Wrapper>

            <HandleBackendService
              templates={templates}
              show={showHandleBackendService}
              setShow={setShowHanldeBackendService}
            />
          </>
        );
      }}
    </LoadingComp>
  );
};

export default BackendServices;
