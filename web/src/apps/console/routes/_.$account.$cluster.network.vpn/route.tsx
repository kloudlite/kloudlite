import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import { IDevices } from '~/console/server/gql/queries/vpn-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { DIALOG_TYPE } from '~/console/utils/commons';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
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
  const [showHandleDevice, setShowHandleDevice] =
    useState<IShowDialog<ExtractNodeType<IDevices> | null>>(null);
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp data={promise}>
      {({ devicesData }) => {
        const devices = devicesData.edges?.map(({ node }) => node);

        console.log(devices);

        return (
          <>
            <Wrapper
              secondaryHeader={{
                title: 'VPN',
                action: devices.length > 0 && (
                  <Button
                    content="Add device"
                    prefix={<Plus />}
                    variant="primary"
                    onClick={() => {
                      setShowHandleDevice({
                        type: DIALOG_TYPE.ADD,
                        data: null,
                      });
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
                    setShowHandleDevice({ type: DIALOG_TYPE.ADD, data: null });
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
            <HandleDevices
              show={showHandleDevice}
              setShow={setShowHandleDevice}
            />
          </>
        );
      }}
    </LoadingComp>
  );
};

export default VPN;
