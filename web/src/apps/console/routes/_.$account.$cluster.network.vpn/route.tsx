import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import DeviceResources from './devices-resources';
import HandleDevices from './handle-devices';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { cluster } = ctx.params;
    const { data, errors } = await GQLServerHandler(ctx.request).listVpnDevices(
      {
        clusterName: cluster,
        pq: getPagination(ctx),
        search: getSearch(ctx),
      }
    );
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
const VPN = () => {
  const [visible, setVisible] = useState(false);
  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          devicesData: fake.ConsoleListVpnDevicesQuery
            .infra_listVPNDevices as any,
        }}
      >
        {({ devicesData }) => {
          const devices = devicesData.edges?.map(({ node }) => node);

          return (
            <Wrapper
              secondaryHeader={{
                title: 'VPN',
                action: devices.length > 0 && (
                  <Button
                    content="Add device"
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
                  content: 'Add device',
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
              <DeviceResources items={devices} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleDevices
        {...{
          isUpdate: false,
          visible,
          setVisible,
        }}
      />
    </>
  );
};

export default VPN;
