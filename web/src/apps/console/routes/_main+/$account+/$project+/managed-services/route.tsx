import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import Tools from './tools';
import BackendServicesResources from './backend-services-resources';
import fake from "~/root/fake-data-generator/fake";

export const loader = (ctx: IRemixCtx) => {
  const { project } = ctx.params;
  const promise = pWrapper(async () => {
    const { data: mData, errors: mErrors } = await GQLServerHandler(
      ctx.request
    ).listProjectMSvs({
      projectName: project,
    });

    const { data: msvTemplates, errors: msvError } = await GQLServerHandler(
      ctx.request
    ).listMSvTemplates({});

    if (mErrors) {
      throw mErrors[0];
    }

    if (msvError) {
      throw msvError[0];
    }

    return { managedServices: mData, templates: msvTemplates };
  });
  return defer({ promise });
};

const KlOperatorServices = () => {
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp
        data={promise}
        skeletonData={{
          managedServices: fake.ConsoleListProjectMSvsQuery.core_listProjectManagedServices as any,
          templates: fake.ConsoleListMSvTemplatesQuery.infra_listManagedServiceTemplates as any
        }}
    >
      {({ managedServices, templates: templatesData }) => {
        const backendServices = parseNodes(managedServices);

        return (
          <Wrapper
            header={{
              title: 'Managed services',
              action: backendServices.length > 0 && (
                <Button
                  variant="primary"
                  content="Create managed service"
                  prefix={<PlusFill />}
                  to="../new-managed-service"
                  LinkComponent={Link}
                />
              ),
            }}
            empty={{
              is: backendServices.length === 0,
              title: 'This is where you’ll manage your Managed services.',
              content: (
                <p>
                  You can create a new backing service and manage the listed
                  backing service.
                </p>
              ),
              action: {
                content: 'Create new managed service',
                prefix: <Plus />,
                to: '../new-managed-service',
                LinkComponent: Link,
              },
            }}
            tools={<Tools />}
          >
            <BackendServicesResources
              items={backendServices}
              templates={templatesData}
            />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default KlOperatorServices;
