/* eslint-disable react/no-unescaped-entities */
import { useState } from 'react';
import { Link } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import AlertDialog from '~/console/components/alert-dialog';
import Wrapper from '~/console/components/wrapper';
import ResourceList from '../../components/resource-list';
import { dummyData } from '../../dummy/data';
import Resources from './resources';
import Filters from './filters';
import Tools from './tools';
import HandleDevice, { ShowQR, ShowWireguardConfig } from './handle-device';

const Vpn = () => {
  const [appliedFilters, setAppliedFilters] = useState(
    dummyData.appliedFilters
  );
  const [currentPage, _setCurrentPage] = useState(1);
  const [itemsPerPage, _setItemsPerPage] = useState(15);
  const [totalItems, _setTotalItems] = useState(100);
  const [showHandleNodePool, setHandleNodePool] = useState(null);
  const [showQRCode, setShowQRCode] = useState(false);
  const [showWireGuardConfig, setShowWireGuardConfig] = useState(false);
  const [showStopNodePool, setShowStopNodePool] = useState(false);
  const [showDeleteNodePool, setShowDeleteNodePool] = useState(false);

  const [data, _setData] = useState(dummyData.devices);

  return (
    <>
      <Wrapper
        header={{
          title: 'VPN',
          action: data.length > 0 && (
            <Button
              variant="primary"
              content="Create new device"
              prefix={PlusFill}
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
            prefix: Plus,
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
        <div className="flex flex-col">
          <Tools />

          <Filters
            appliedFilters={appliedFilters}
            setAppliedFilters={setAppliedFilters}
          />
        </div>
        <div className="flex flex-col gap-lg">
          <div className="bodyLg-medium text-text-strong">Personal Device</div>
          <ResourceList>
            {data
              .filter((d) => d.category === 'personal')
              .map((d) => (
                <ResourceList.ResourceItem key={d.id} textValue={d.id}>
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
                    onStop={(e) => {
                      setShowStopNodePool(e);
                    }}
                    onDelete={(e) => {
                      setShowDeleteNodePool(e);
                    }}
                  />
                </ResourceList.ResourceItem>
              ))}
          </ResourceList>
        </div>
        <div className="flex flex-col gap-lg">
          <div className="bodyLg-medium text-text-strong">Team's Device</div>
          <ResourceList>
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
                    onStop={(e) => {
                      setShowStopNodePool(e);
                    }}
                    onDelete={(e) => {
                      setShowDeleteNodePool(e);
                    }}
                  />
                </ResourceList.ResourceItem>
              ))}
          </ResourceList>
        </div>
      </Wrapper>

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
