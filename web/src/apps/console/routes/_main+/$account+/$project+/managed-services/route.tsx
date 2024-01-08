import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { Dispatch, SetStateAction, useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { IMSvTemplates } from '~/console/server/gql/queries/managed-templates-queries';
import Tools from './tools';
import HandleBackendService from './handle-backend-service';
import BackendServicesResources from './backend-services-resources';

export const loader = (ctx: IRemixCtx) => {
  const { cluster } = ctx.params;
  const promise = pWrapper(async () => {
    const { data: mData, errors: mErrors } = await GQLServerHandler(
      ctx.request
    ).listClusterMSvs({
      clusterName: cluster,
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

const SetValue = ({
  templates,
  setTemplate,
}: {
  templates: IMSvTemplates;
  setTemplate: Dispatch<SetStateAction<IMSvTemplates | null>>;
}) => {
  useEffect(() => {
    setTemplate(templates);
  }, []);
  return null;
};

const KlOperatorServices = () => {
  const [visible, setVisible] = useState(false);
  const { promise } = useLoaderData<typeof loader>();
  const [templates, setTemplates] = useState<IMSvTemplates | null>(null);

  return (
    <>
      <LoadingComp data={promise}>
        {({ managedServices, templates: templatesData }) => {
          const backendServices = parseNodes(managedServices);

          return (
            <>
              <SetValue templates={templatesData} setTemplate={setTemplates} />
              <Wrapper
                header={{
                  title: 'Managed services',
                  action: backendServices.length > 0 && (
                    <Button
                      variant="primary"
                      content="Create managed service"
                      prefix={<PlusFill />}
                      onClick={() => {
                        setVisible(true);
                      }}
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
                    onClick: () => {
                      setVisible(true);
                    },
                  },
                }}
                tools={<Tools />}
              >
                <BackendServicesResources
                  items={backendServices}
                  templates={templatesData}
                />
              </Wrapper>
            </>
          );
        }}
      </LoadingComp>
      <HandleBackendService
        {...{
          isUpdate: false,
          visible,
          setVisible,
          templates: templates || [],
        }}
      />
    </>
  );
};

export default KlOperatorServices;
