import { Plus } from '~/console/components/icons';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { Button } from '~/components/atoms/button.jsx';
import Wrapper from '~/console/components/wrapper';
import { parseNodes } from '~/console/server/r-utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { getPagination, getSearch } from '~/console/server/utils/common';
import fake from '~/root/fake-data-generator/fake';
import Tools from './tools';
import VPNResourcesV2 from './vpn-resources-v2';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listGlobalVpnDevices({
      gvpn: 'default',
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });

    if (errors) {
      throw errors[0];
    }

    return {
      devicesData: data,
    };
  });

  return defer({ promise });
};

const Devices = () => {
  const { promise } = useLoaderData<typeof loader>();

  const getEmptyState = ({ deviceCount }: { deviceCount: number }) => {
    if (deviceCount === 0) {
      return {
        is: true,
        title: 'This is where you’ll manage your devices.',
        content: (
          <p>You can create a new device and manage the listed devices.</p>
        ),
        action: {
          content: 'Create new device',
          prefix: <Plus />,
          // LinkComponent: Link,
        },
      };
    }

    return {
      is: false,
      title: 'This is where you’ll manage your devices.',
      content: (
        <p>You can create a new device and manage the listed devices.</p>
      ),
      action: {
        content: 'Create new device',
        prefix: <Plus />,
        // LinkComponent: Link,
      },
    };
  };

  return (
    <LoadingComp
      data={promise}
      skeletonData={{
        devicesData: fake.ConsoleListGlobalVpnDevicesQuery
          .infra_listGlobalVPNDevices as any,
      }}
    >
      {({ devicesData }) => {
        const vpnDevices = parseNodes(devicesData);

        if (!vpnDevices) {
          return null;
        }

        const { pageInfo, totalCount } = devicesData;
        return (
          <Wrapper
            header={{
              title: 'Vpn Devices',
              action: vpnDevices.length > 0 && (
                <Button
                  content="Create device"
                  variant="primary"
                  prefix={<Plus />}
                  // LinkComponent={Link}
                />
              ),
            }}
            empty={getEmptyState({
              deviceCount: vpnDevices.length,
            })}
            pagination={{
              pageInfo,
              totalCount,
            }}
            tools={<Tools />}
          >
            <VPNResourcesV2 items={vpnDevices} />
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default Devices;
