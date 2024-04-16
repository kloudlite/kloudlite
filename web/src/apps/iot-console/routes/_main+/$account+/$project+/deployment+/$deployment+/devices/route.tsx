import { Plus, PlusFill } from '@jengaicons/react';
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
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import Tools from './tools';
import DeviceResource from './device-resource';
import HandleDevice from './handle-device';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  const { deployment, project } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).listIotDevices(
      {
        projectName: project,
        deploymentName: deployment,
        pq: getPagination(ctx),
        search: getSearch(ctx),
      }
    );
    if (errors) {
      throw errors[0];
    }

    return { devicesData: data };
  });

  return defer({ promise });
};

const Apps = () => {
  const { promise } = useLoaderData<typeof loader>();
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    clearAppState();
  }, []);

  return (
    <>
      <LoadingComp
        data={promise}
        // skeletonData={{
        //   appsData: fake.ConsoleListAppsQuery.core_listApps as any,
        // }}
      >
        {({ devicesData }) => {
          const devices = parseNodes(devicesData);
          if (!devices) {
            return null;
          }

          return (
            <div>
              <Wrapper
                header={{
                  title: 'Devices',
                  action: devices?.length > 0 && (
                    <Button
                      variant="primary"
                      content="Create new device"
                      prefix={<PlusFill />}
                      onClick={() => {
                        setVisible(true);
                      }}
                      // to="../new-app"
                      // LinkComponent={Link}
                    />
                  ),
                }}
                empty={{
                  is: devices?.length === 0,
                  title: 'This is where you’ll manage your Devices.',
                  content: (
                    <p>
                      You can create a new device and manage the listed device.
                    </p>
                  ),
                  action: {
                    content: 'Create new device',
                    prefix: <Plus />,
                    onClick: () => {
                      setVisible(true);
                    },
                    LinkComponent: Link,
                    //   LinkComponent: Link,
                    //   to: '../new-app',
                  },
                }}
                tools={<Tools />}
              >
                <DeviceResource items={devices} />
              </Wrapper>
            </div>
          );
        }}
      </LoadingComp>
      <HandleDevice
        {...{
          visible,
          setVisible,
          isUpdate: false,
        }}
      />
    </>
  );
};

export default Apps;
