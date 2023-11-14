import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData, useOutletContext } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  getScopeAndProjectQuery,
  parseNodes,
} from '~/console/server/r-utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { IWorkspaceContext } from '../_.$account.$cluster.$project.$scope.$workspace/route';
import BackendServicesResources from './backend-services-resources';
import HandleBackendService from './handle-backend-service';
import Tools from './tools';

export const loader = (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    const { data: mData, errors: mErrors } = await GQLServerHandler(
      ctx.request
    ).listManagedServices({ ...getScopeAndProjectQuery(ctx) });

    if (mErrors) {
      throw mErrors[0];
    }

    return { managedServices: mData };
  });
  return defer({ promise });
};

const BackendServices = () => {
  const [showHandleBackendService, setShowHanldeBackendService] =
    useState<IShowDialog>(null);
  const { promise } = useLoaderData<typeof loader>();

  const { managedTemplates } = useOutletContext<IWorkspaceContext>();

  return (
    <LoadingComp data={promise}>
      {({ managedServices }) => {
        const backendServices = parseNodes(managedServices);

        return (
          <>
            <Wrapper
              header={{
                title: 'Backing services',
                action: backendServices.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create backing service"
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
                  },
                },
              }}
            >
              <Tools />
              <BackendServicesResources
                items={backendServices}
                templates={managedTemplates}
              />
            </Wrapper>

            <HandleBackendService
              templates={managedTemplates}
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
