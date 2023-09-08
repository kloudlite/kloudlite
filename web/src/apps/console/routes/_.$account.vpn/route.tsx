import { useState } from 'react';
import { Link, useLoaderData, useOutletContext } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import AlertDialog from '~/console/components/alert-dialog';
import Wrapper from '~/console/components/wrapper';
import { defer } from '@remix-run/node';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { IRemixCtx } from '~/root/lib/types/common';
import {
  getPagination,
  getSearch,
  listOrGrid,
} from '~/console/server/utils/common';
import { parseNodes } from '~/console/server/r-urils/common';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import ResourceList from '../../components/resource-list';
import { dummyData } from '../../dummy/data';
import Resources from './resources';
import Tools from './tools';
import HandleDevice, { ShowQR, ShowWireguardConfig } from './handle-device';
import { IConsoleRootContext } from '../_';

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
  return defer({ promise });
};

const Vpn = () => {
  const { user } = useOutletContext<IConsoleRootContext>();
  const [viewMode, setViewMode] = useState<listOrGrid>('list');
  const [currentPage, _setCurrentPage] = useState(1);
  const [itemsPerPage, _setItemsPerPage] = useState(15);
  const [totalItems, _setTotalItems] = useState(100);
  const [showHandleNodePool, setHandleNodePool] = useState<{
    type: 'add' | 'edit';
    data: any;
  } | null>(null);
  const [showQRCode, setShowQRCode] = useState(false);
  const [showWireGuardConfig, setShowWireGuardConfig] = useState(false);
  const [showStopNodePool, setShowStopNodePool] = useState(false);
  const [showDeleteNodePool, setShowDeleteNodePool] = useState(false);

  const [data, _setData] = useState(dummyData.devices);
  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp data={promise}>
        {({ vpnsData }) => {
          const devices = parseNodes(vpnsData);
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
            >
              <Tools viewMode={viewMode} setViewMode={setViewMode} />

              <div className="flex flex-col gap-lg">
                <div className="bodyLg-medium text-text-strong">
                  Personal Device
                </div>
                <ResourceList mode={viewMode}>
                  {devices
                    .filter((d) => d.createdBy.userId === user.id)
                    .map((d) => (
                      <ResourceList.ResourceItem
                        key={d.metadata.name}
                        textValue={d.metadata.name}
                      >
                        <Resources
                          item={d}
                          onEdit={() => {
                            setHandleNodePool({ type: 'edit', data: null });
                          }}
                          onQR={() => {
                            setShowQRCode(true);
                          }}
                          onWireguard={() => {
                            setShowWireGuardConfig(true);
                          }}
                          onStop={(e: any) => {
                            setShowStopNodePool(e);
                          }}
                          onDelete={(e: any) => {
                            setShowDeleteNodePool(e);
                          }}
                        />
                      </ResourceList.ResourceItem>
                    ))}
                </ResourceList>
              </div>
              <div className="flex flex-col gap-lg">
                <div className="bodyLg-medium text-text-strong">
                  Team&apos;s Device
                </div>
                <ResourceList mode={viewMode}>
                  {data
                    .filter((d) => d.category === 'team')
                    .map((d) => (
                      <ResourceList.ResourceItem key={d.id} textValue={d.id}>
                        <Resources
                          item={d}
                          onEdit={() => {
                            setHandleNodePool({ type: 'add', data: null });
                          }}
                          onQR={() => {
                            setShowQRCode(true);
                          }}
                          onWireguard={() => {
                            setShowWireGuardConfig(true);
                          }}
                          onStop={(e: any) => {
                            setShowStopNodePool(e);
                          }}
                          onDelete={(e: any) => {
                            setShowDeleteNodePool(e);
                          }}
                        />
                      </ResourceList.ResourceItem>
                    ))}
                </ResourceList>
              </div>
            </Wrapper>
          );
        }}
      </LoadingComp>

      <HandleDevice show={showHandleNodePool} setShow={setHandleNodePool} />

      <ShowQR show={showQRCode} setShow={setShowQRCode} />
      <ShowWireguardConfig
        show={showWireGuardConfig}
        setShow={setShowWireGuardConfig}
      />

      <AlertDialog
        show={showStopNodePool}
        setShow={setShowStopNodePool}
        title="Stop nodepool"
        message={"Are you sure you want to stop 'kloud-root-ca.crt'?"}
        type="warning"
        okText="Stop"
        onSubmit={(e) => {
          console.log(e);
        }}
      />
      <AlertDialog
        show={showDeleteNodePool}
        setShow={setShowDeleteNodePool}
        title="Delete nodepool"
        message={"Are you sure you want to delete 'kloud-root-ca.crt'?"}
        type="critical"
        okText="Delete"
        onSubmit={(e) => {
          console.log(e);
        }}
      />
    </>
  );
};

export default Vpn;
