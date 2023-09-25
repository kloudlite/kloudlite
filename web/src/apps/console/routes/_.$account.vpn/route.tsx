import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wip from '~/console/components/wip';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { dummyData } from '../../dummy/data';
import HandleDevice, { ShowQR, ShowWireguardConfig } from './handle-device';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureClusterSet(ctx);
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(ctx.request).listVpnDevices(
      {
        pq: getPagination(ctx),
        search: getSearch(ctx),
      }
    );
    if (errors) {
      throw errors[0];
    }

    return { vpnsData: data };
  });

  const clusterPromise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(ctx.request).listClusters({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      throw errors[0];
    }

    return { clustersData: data };
  });
  return defer({ promise, clusterPromise });
};

const Vpn = () => {
  const [currentPage, _setCurrentPage] = useState(1);
  const [itemsPerPage, _setItemsPerPage] = useState(15);
  const [totalItems, _setTotalItems] = useState(100);
  const [showHandleNodePool, setHandleNodePool] = useState<IShowDialog>(null);
  const [showQRCode, setShowQRCode] = useState<IShowDialog>(null);
  const [showWireGuardConfig, setShowWireGuardConfig] =
    useState<IShowDialog>(null);

  const [data, _setData] = useState(dummyData.devices);
  const { promise, clusterPromise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp data={promise}>
        {({ vpnsData }) => {
          const _devices = parseNodes(vpnsData);
          return (
            <Wrapper
              header={{
                title: 'VPN',
                action: data.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create new device"
                    prefix={<PlusFill />}
                    onClick={() => {
                      setHandleNodePool({ type: 'add', data: null });
                    }}
                  />
                ),
              }}
              empty={{
                is: data.length === 0,
                title:
                  'This is the place where you will handle and oversee your VPN.',
                content: (
                  <p>
                    You have the option to include a new VPN and oversee the
                    management of existing listed VPN.
                  </p>
                ),
                action: {
                  content: 'Create new device',
                  prefix: <Plus />,
                  LinkComponent: Link,
                  onClick: () => {
                    setHandleNodePool({ type: 'add', data: null });
                  },
                },
              }}
              pagination={{
                currentPage,
                itemsPerPage,
                totalItems,
              }}
              tools={<Tools />}
            >
              <Wip />
            </Wrapper>
          );
        }}
      </LoadingComp>

      <LoadingComp skeleton={<span />} data={clusterPromise}>
        {({ clustersData }) => {
          const clusters = parseNodes(clustersData);
          return (
            <HandleDevice
              clusters={clusters}
              show={showHandleNodePool}
              setShow={setHandleNodePool}
            />
          );
        }}
      </LoadingComp>

      <ShowQR show={showQRCode} setShow={setShowQRCode} />
      <ShowWireguardConfig
        show={showWireGuardConfig}
        setShow={setShowWireGuardConfig}
      />
    </>
  );
};

export default Vpn;
