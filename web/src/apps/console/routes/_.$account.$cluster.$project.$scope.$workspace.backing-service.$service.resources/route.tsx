import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData, useOutletContext } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import { parseNodes } from '~/console/server/r-utils/common';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { getManagedTemplate } from '~/console/utils/commons';
import Wip from '~/root/lib/client/components/wip';
import { IRemixCtx } from '~/root/lib/types/common';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '../../server/utils/auth-utils';
import { IManagedServiceContext } from '../_.$account.$cluster.$project.$scope.$workspace.backing-service.$service/route';
import HandleBackendResources from './handle-backend-resources';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { project } = ctx.params;
  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listWorkspaces(
      {
        project: {
          type: 'name',
          value: project,
        },
        pagination: getPagination(ctx),
        search: getSearch(ctx),
      }
    );
    if (errors) {
      throw errors[0];
    }
    console.log(JSON.stringify(data, null, 2));

    return {
      workspacesData: data || {},
    };
  });

  return defer({ promise });
};

const BackingResources = () => {
  const [showBackendResourceDialog, setShowBackendResourceDialog] =
    useState<IShowDialog>(null);
  const { promise } = useLoaderData<typeof loader>();

  const { managedTemplates, backendService } =
    useOutletContext<IManagedServiceContext>();

  return (
    <>
      <LoadingComp data={promise}>
        {({ workspacesData }) => {
          const workspaces = parseNodes(workspacesData);

          if (!workspaces) {
            return null;
          }

          return (
            <Wrapper
              header={{
                title: 'Resources',
                action: (
                  <Button
                    variant="primary"
                    content="Create resource"
                    prefix={<PlusFill />}
                    onClick={() => {
                      setShowBackendResourceDialog({ type: '', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                is: workspaces.length === 0,
                title:
                  'This is where you’ll manage your backing service resources..',
                content: (
                  <p>
                    You can create a new resource and manage the listed
                    resources.
                  </p>
                ),
                action: {
                  content: 'Create new resource',
                  prefix: <Plus />,
                  onClick: () => {
                    setShowBackendResourceDialog({ type: '', data: null });
                  },
                },
              }}
              tools={<Tools />}
            >
              <Wip />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleBackendResources
        template={
          getManagedTemplate({
            templates: managedTemplates,
            kind: backendService.spec.msvcKind.kind || '',
            apiVersion: backendService.spec.msvcKind.apiVersion,
          })!
        }
        setShow={setShowBackendResourceDialog}
        show={showBackendResourceDialog}
      />
    </>
  );
};

export default BackingResources;
