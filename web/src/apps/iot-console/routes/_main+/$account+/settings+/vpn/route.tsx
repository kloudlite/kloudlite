import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import logger from '~/lib/client/helpers/log';
import { IRemixCtx } from '~/lib/types/common';
import { getPagination, getSearch } from '~/console/server/utils/common';
import HandleConsoleDevices from '~/console/page-components/handle-console-devices';
import fake from '~/root/fake-data-generator/fake';
import Tools from './tools';
import ConsoleDeviceResourcesV2 from './console-device-resources-v2';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listConsoleVpnDevices({
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      devicesData: data || {},
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
          devicesData: fake.ConsoleListConsoleVpnDevicesQuery
            .core_listVPNDevices as any,
        }}
      >
        {({ devicesData }) => {
          const devices = devicesData?.edges?.map(({ node }) => node);
          if (!devices) {
            return null;
          }
          return (
            <Wrapper
              secondaryHeader={{
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
              <ConsoleDeviceResourcesV2 items={devices} />
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
