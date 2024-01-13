import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import Tools from './tools';
import ConsoleDeviceResources from './console-devices-resources';
import HandleConsoleDevices from './handle-console-devices';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listConsoleVpnDevicesForUser({});
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      devicesData: data || [],
    };
  });

  return defer({ promise });
};
const ConsoleVPN = () => {
  const [visible, setVisible] = useState(false);
  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          devicesData: fake.ConsoleListConsoleVpnDevicesForUserQuery
            .core_listVPNDevicesForUser as any,
        }}
      >
        {({ devicesData }) => {
          const devices = devicesData;
          return (
            <Wrapper
              header={{
                title: 'VPN devices',
                action: devices.length > 0 && (
                  <Button
                    content="Create Device"
                    prefix={<Plus />}
                    variant="primary"
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                is: devices.length === 0,
                action: {
                  content: 'Create Device',
                  prefix: <Plus />,
                  variant: 'primary',
                  onClick: () => {
                    setVisible(true);
                  },
                },
                title:
                  'This is the place where you will handle and oversee your VPN.',
                content: (
                  <p>
                    You have the option to include a new VPN and oversee the
                    management of existing listed VPN.
                  </p>
                ),
              }}
              tools={<Tools />}
            >
              <ConsoleDeviceResources items={devices} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleConsoleDevices
        {...{
          isUpdate: false,
          visible,
          setVisible,
        }}
      />
    </>
  );
};

export default ConsoleVPN;
